package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"compress/gzip"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"scraper/internal/scraper"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"

	"github.com/google/brotli/go/cbrotli"
)

const (
	defaultTransPerMinute = 5.0
	siteDefaultMaxWorkers = 2
	defaultMaxWorkers     = 100
)

var CustomWebsites = []struct {
	site        string
	workers     int64
	transPerMin int64
}{
	{"www.thekitchn.com", 1, 2},
	{"www.beefitswhatsfordinner.com", 1, 5},
	{"www.tasteofhome.com", 5, 300},
}

type perSite struct {
	sourceURL             string
	ctx                   context.Context
	limit                 rate.Limit
	limiter               *rate.Limiter
	maxWorkers            *semaphore.Weighted
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
	tr.Proxy = http.ProxyFromEnvironment
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

	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Error(err)
	}
	return &SiteClient{
		ReplyChan:   make(chan *scraper.RecipeObject),
		perSiteLock: &sync.RWMutex{},
		site:        make(map[string]*perSite),
		scrape:      scraper.New(),
		maxWorkers:  semaphore.NewWeighted(defaultMaxWorkers),
		Client:      &http.Client{Jar: jar, Transport: tr, Timeout: time.Second * 60},
	}
}

func (x *SiteClient) newSite(sourceURL string) *perSite {
	out := &perSite{
		sourceURL: sourceURL,
		ctx:       context.Background(),
		limit:     rate.Limit(defaultTransPerMinute),
		queue:     make(chan string, 1000),
		queueSem:  sync.NewCond(&sync.RWMutex{}),
		parent:    x,
	}
	u, err := url.Parse(sourceURL)
	if err == nil {
		for _, site := range CustomWebsites {
			if u.Host == site.site {
				out.maxWorkers = semaphore.NewWeighted(site.workers)
				// 10 transactions/min
				out.limiter = rate.NewLimiter(rate.Limit(float64(site.transPerMin)/60.0), 1)
				break
			}
		}
		out.maxWorkers = semaphore.NewWeighted(siteDefaultMaxWorkers)
		out.limiter = rate.NewLimiter(rate.Limit(defaultTransPerMinute/60.0), 1)
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
				SourceURL:  sourceURL,
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
				SourceURL:  sourceURL,
				StatusCode: http.StatusServiceUnavailable,
				Error:      x.transportErrorMessage,
			}
			x.parent.ReplyChan <- recipe
			continue
		}

		// limit the number of concurrent transactions
		x.maxWorkers.Acquire(x.ctx, 1)
		// limit the rate on the website
		//x.limiter.Wait(x.ctx)

		go func(x *perSite, sourceURL string) {
			var resp *http.Response
			var err error
			defer x.maxWorkers.Release(1)
			for retry := 0; retry < 2; retry++ {
				req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
				if err != nil {
					return
				}
				req.Header.Set("User-Agent", RandomUserAgent())
				req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
				//req.Header.Set("accept","*/*")
				req.Header.Set("referer", sourceURL)
				req.Header.Add("Accept-Encoding", "gzip")
				req.Header.Add("Accept-Encoding", "br")
				if u, err := url.Parse(sourceURL); err != nil {
					req.Header.Set("host", u.Host)
				}

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
							SourceURL:  sourceURL,
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
				SourceURL:  sourceURL,
				StatusCode: resp.StatusCode,
			}
			if err != nil {
				recipe.Error = err.Error()
			}
			x.parent.ReplyChan <- recipe
		}(x, sourceURL)
	}
}
