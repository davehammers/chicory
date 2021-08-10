package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"scraper/internal/scraper"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type perSite struct {
	siteURL                 string
	ctx                     context.Context
	limiter                 *rate.Limiter
	queueSem                *sync.Cond
	queue                   []string // queue URLs to be processed for this site
	requestCount            int
	transportErrorCount     int
	transportErrorMessage   string
	parent                  *siteClient
}

type siteClient struct {
	ReplyChan   chan *scraper.RecipeObject
	perSiteLock *sync.RWMutex
	site        map[string]*perSite
	scrape      *scraper.Scraper
	maxWorkers  *semaphore.Weighted
	client      *http.Client
}

func NewSiteClient() *siteClient {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 500
	tr.IdleConnTimeout = 30 * time.Second
	tr.MaxConnsPerHost = 2
	tr.MaxIdleConnsPerHost = 2
	tr.ResponseHeaderTimeout = 45 * time.Second
	tr.DisableKeepAlives = true
	tr.TLSHandshakeTimeout=   700 * time.Millisecond
	tr.TLSClientConfig.InsecureSkipVerify = true
	return &siteClient{
		ReplyChan:   make(chan *scraper.RecipeObject),
		perSiteLock: &sync.RWMutex{},
		site:        make(map[string]*perSite),
		scrape:      &scraper.Scraper{},
		maxWorkers:  semaphore.NewWeighted(100),
		client:      &http.Client{Transport: tr, Timeout: time.Second * 30},
	}
}

func (x *siteClient) newSite(siteURL string) *perSite {
	out := &perSite{
		siteURL:  siteURL,
		ctx:      context.Background(),
		limiter:  rate.NewLimiter(rate.Limit(1), 1),
		queue:    make([]string, 0),
		queueSem: sync.NewCond(&sync.RWMutex{}),
		parent:   x,
	}
	go out.queueHandler()

	return out
}

// GetRecipe - extract any recipes from the URL site provided. returns interface
func (x *siteClient) SiteGetRecipe(siteURL string) (err error) {
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
	site.queue = append(site.queue, siteURL)
	//site.queueSem.Signal()
	return
}

func (x *perSite) queueHandler() {
	var resp *http.Response
	var err error
	for {
		// pull URL off the queue
		if len(x.queue) == 0 {
			// TODO sync with sender
			time.Sleep(time.Second)
			continue
		}
		siteURL := x.queue[0]
		x.queue = x.queue[1:]
		x.requestCount++
		// is URL parsable
		if _, err = url.Parse(siteURL); err != nil {
			recipe := &scraper.RecipeObject{
				SiteURL:    siteURL,
				StatusCode: http.StatusBadRequest,
				Error:      err.Error(),
			}
			x.parent.ReplyChan <- recipe
			continue
		}
		//if x.requestCount > 1000 {
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
		if err = x.limiter.Wait(x.ctx); err != nil {
			log.Warn(err)
		}

		x.parent.maxWorkers.Acquire(x.ctx, 1)
		resp, err = x.parent.client.Get(siteURL)
		x.parent.maxWorkers.Release(1)

		switch err {
		case nil:
			// successfully communicated with a domain name
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				resp.StatusCode = http.StatusInternalServerError
			} else if len(body) == 0 {
				resp.StatusCode = http.StatusLengthRequired
				x.transportErrorCount++
				x.transportErrorMessage = "Website closed connection before any data sent"
			}
			switch resp.StatusCode {
			case http.StatusOK:
				x.transportErrorCount = 0
				x.transportErrorMessage = ""
				recipe, _ := x.parent.scrape.ScrapeRecipe(siteURL, body)
				x.parent.ReplyChan <- recipe
			default:
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
			code := http.StatusServiceUnavailable
			if strings.Contains(err.Error(), "time") {
				code = http.StatusGatewayTimeout
			}
			recipe := &scraper.RecipeObject{
				SiteURL:    siteURL,
				StatusCode: code,
				Error:      err.Error(),
			}
			x.parent.ReplyChan <- recipe
		}
	}
}
