package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"scraper/internal/scraper"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"

	"github.com/google/brotli/go/cbrotli"
)

const (
	defaultConcurrentPerSite = .1
	siteDefaultMaxWorkers = 4
	defaultMaxWorkers        = 100
)

type perSite struct {
	sourceURL               string
	ctx                   context.Context
	limit rate.Limit
	limiter               *rate.Limiter
	maxWorkers  *semaphore.Weighted
	queueSem              *sync.Cond
	queue                 chan string
	requestCount          int
	transportErrorCount   int
	transportErrorMessage string
	parent                *SiteClient
}

type SiteClient struct {
	ReplyChan   chan *scraper.RecipeObject
	perSiteLock *sync.RWMutex
	site        map[string]*perSite
	scrape      *scraper.Scraper
	maxWorkers  *semaphore.Weighted
	Client      *http.Client
}

func NewSiteClient() *SiteClient {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 500
	tr.IdleConnTimeout = 30 * time.Second
	tr.MaxConnsPerHost = 2
	tr.MaxIdleConnsPerHost = 2
	tr.ResponseHeaderTimeout = 60 * time.Second
	tr.DisableKeepAlives = true
	tr.TLSClientConfig.InsecureSkipVerify = true
	// We use ABSURDLY large keys, and should probably not.
	tr.TLSHandshakeTimeout = 45 * time.Second
	tr.DisableCompression = false

	return &SiteClient{
		ReplyChan:   make(chan *scraper.RecipeObject),
		perSiteLock: &sync.RWMutex{},
		site:        make(map[string]*perSite),
		scrape:      scraper.New(),
		maxWorkers:  semaphore.NewWeighted(defaultMaxWorkers),
		Client:      &http.Client{Transport: tr, Timeout: time.Second * 60},
	}
}

func (x *SiteClient) newSite(sourceURL string) *perSite {
	out := &perSite{
		sourceURL:  sourceURL,
		ctx:      context.Background(),
		limit: rate.Limit(defaultConcurrentPerSite),
		limiter:  rate.NewLimiter(rate.Limit(defaultConcurrentPerSite), 1),
		maxWorkers:  semaphore.NewWeighted(siteDefaultMaxWorkers),
		queue:    make(chan string, 1000),
		queueSem: sync.NewCond(&sync.RWMutex{}),
		parent:   x,
	}
	go out.queueHandler()

	return out
}

// GetRecipe - extract any recipes from the URL site provided. returns interface
func (x *SiteClient) SiteGetRecipe(sourceURL string) (err error) {
	u, err := url.Parse(sourceURL)
	if err != nil {
		fmt.Println("host", err)
		return
	}
	site, ok := x.site[u.Host]
	if !ok {
		x.perSiteLock.Lock()
		site = x.newSite(sourceURL)
		x.site[u.Host] = site
		x.perSiteLock.Unlock()
	}
	go func(sourceURL string) {
		site.queue <- sourceURL
	}(sourceURL)
	return
}

func (x *perSite) queueHandler() {
	for sourceURL := range x.queue {
		x.requestCount++
		// is URL parsable
		if _, err := url.Parse(sourceURL); err != nil {
			recipe := &scraper.RecipeObject{
				SourceURL:    sourceURL,
				StatusCode: http.StatusBadRequest,
				Error:      err.Error(),
			}
			x.parent.ReplyChan <- recipe
			continue
		}
		//if x.requestCount > 10 {
		//	// silently skip requests
		//	continue
		//}
		if x.transportErrorCount > 50 {
			recipe := &scraper.RecipeObject{
				SourceURL:    sourceURL,
				StatusCode: http.StatusServiceUnavailable,
				Error:      x.transportErrorMessage,
			}
			x.parent.ReplyChan <- recipe
			continue
		}

		// limit the number of concurrent transactions
		x.maxWorkers.Acquire(x.ctx, 1)

		go func(x *perSite, sourceURL string) {
			var resp *http.Response
			var err error
			defer x.maxWorkers.Release(1)
			for retry := 0; retry < 2; retry++ {
				req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36")
				req.Header.Set("accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
				req.Header.Set("referer", sourceURL)
				req.Header.Add("Accept-Encoding", "gzip")
				req.Header.Add("Accept-Encoding", "br")


				x.parent.maxWorkers.Acquire(x.ctx, 1)
				resp, err = x.parent.Client.Do(req)
				x.parent.maxWorkers.Release(1)

				switch err {
				case nil:
					var body []byte
					// successfully communicated with a domain name
					defer resp.Body.Close()
					switch resp.Header.Get("Content-Encoding") {
					case "br":
						// some web sites compress data even if not requested
						reader := cbrotli.NewReader(resp.Body)
						body, err = ioutil.ReadAll(reader)
						reader.Close()
					case "gzip":
						// some web sites compress data even if not requested
						reader, err := gzip.NewReader(resp.Body)
						if err != nil {
							resp.StatusCode = http.StatusInternalServerError
							continue
						}
						body, err = ioutil.ReadAll(reader)
						reader.Close()
					default:
						body, err = ioutil.ReadAll(resp.Body)
					}

					if err != nil {
						resp.StatusCode = http.StatusInternalServerError
						continue
					}

					switch resp.StatusCode {
					case http.StatusOK:
						if len(body) == 0 {
							resp.StatusCode = http.StatusLengthRequired
							x.transportErrorCount++
							x.transportErrorMessage = "Website closed connection before any data sent"
							continue
						}
						x.transportErrorCount = 0
						x.transportErrorMessage = ""
						recipe, _ := x.parent.scrape.ScrapeRecipe(sourceURL, body)
						x.parent.ReplyChan <- recipe
						return
					case http.StatusForbidden:
						continue
					case http.StatusLengthRequired:
						continue
					default:
						recipe := &scraper.RecipeObject{
							SourceURL:    sourceURL,
							StatusCode: resp.StatusCode,
							Error:      "HTTP error",
						}
						x.parent.ReplyChan <- recipe
						return
					}
				default:
					// Transport failure
					x.transportErrorCount++
					x.transportErrorMessage = err.Error()
					if resp == nil {
						resp = &http.Response{}
					}
					resp.StatusCode = http.StatusServiceUnavailable
					err = fmt.Errorf("HTTP code")
				}
			}
			recipe := &scraper.RecipeObject{
				SourceURL:    sourceURL,
				StatusCode: resp.StatusCode,
			}
			if err != nil {
				recipe.Error = err.Error()
			}
			x.parent.ReplyChan <- recipe
		}(x, sourceURL)
	}
}
