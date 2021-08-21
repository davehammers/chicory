package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"scraper/internal/scraper"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

const (
	defaultConcurrentPerSite = .1
	defaultMaxWorkers        = 100
)

type perSite struct {
	siteURL               string
	ctx                   context.Context
	limit rate.Limit
	limiter               *rate.Limiter
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

func (x *SiteClient) newSite(siteURL string) *perSite {
	out := &perSite{
		siteURL:  siteURL,
		ctx:      context.Background(),
		limit: rate.Limit(defaultConcurrentPerSite),
		limiter:  rate.NewLimiter(rate.Limit(defaultConcurrentPerSite), 1),
		queue:    make(chan string, 1000),
		queueSem: sync.NewCond(&sync.RWMutex{}),
		parent:   x,
	}
	go out.queueHandler()

	return out
}

// GetRecipe - extract any recipes from the URL site provided. returns interface
func (x *SiteClient) SiteGetRecipe(siteURL string) (err error) {
	u, err := url.Parse(siteURL)
	if err != nil {
		fmt.Println("host", err)
		return
	}
	site, ok := x.site[u.Host]
	if !ok {
		x.perSiteLock.Lock()
		site = x.newSite(siteURL)
		x.site[u.Host] = site
		x.perSiteLock.Unlock()
	}
	go func(siteURL string) {
		site.queue <- siteURL
	}(siteURL)
	return
}

func (x *perSite) queueHandler() {
	for siteURL := range x.queue {
		x.requestCount++
		// is URL parsable
		if _, err := url.Parse(siteURL); err != nil {
			recipe := &scraper.RecipeObject{
				SiteURL:    siteURL,
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
				SiteURL:    siteURL,
				StatusCode: http.StatusServiceUnavailable,
				Error:      x.transportErrorMessage,
			}
			x.parent.ReplyChan <- recipe
			continue
		}

		// rate limiter for the domain to avoid overloading website
		if err := x.limiter.Wait(x.ctx); err != nil {
			log.Warn(err)
		}

		go func(x *perSite, siteURL string) {
			var resp *http.Response
			var err error
			for retry := 0; retry < 2; retry++ {
				x.parent.maxWorkers.Acquire(x.ctx, 1)
				resp, err = x.parent.Client.Get(siteURL)
				x.parent.maxWorkers.Release(1)

				switch err {
				case nil:
					// successfully communicated with a domain name
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						resp.StatusCode = http.StatusInternalServerError
						continue
					}
					resp.Body.Close()

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
						recipe, _ := x.parent.scrape.ScrapeRecipe(siteURL, body)
						x.parent.ReplyChan <- recipe
						if x.limit < 20.0 {
							x.limit += .1
							x.limiter.SetLimit(x.limit)
						}
						return
					case http.StatusForbidden:
						continue
					case http.StatusLengthRequired:
						continue
					default:
						if x.limit > .4 {
							x.limit -= .1
							x.limiter.SetLimit(x.limit)
						}
						recipe := &scraper.RecipeObject{
							SiteURL:    siteURL,
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
				SiteURL:    siteURL,
				StatusCode: resp.StatusCode,
			}
			if err != nil {
				recipe.Error = err.Error()
			}
			x.parent.ReplyChan <- recipe
		}(x, siteURL)
	}
}
