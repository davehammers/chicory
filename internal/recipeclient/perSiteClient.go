package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"context"
	"fmt"
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"scraper/internal/scraper"
	"sync"
	"time"
)

type perSite struct {
	siteURL  string
	client   *fasthttp.Client
	ctx      context.Context
	limiter  *rate.Limiter
	cache    *ristretto.Cache
	queueSem *sync.Cond
	queue    []*string // queue URLs to be processed for this site
	parent   *siteClient
}

type siteClient struct {
	ReplyChan   chan *scraper.RecipeObject
	perSiteLock *sync.RWMutex
	site        map[string]*perSite
	scrape      *scraper.Scraper
	maxWorkers *semaphore.Weighted
}

func NewSiteClient() *siteClient {
	return &siteClient{
		ReplyChan:   make(chan *scraper.RecipeObject),
		perSiteLock: &sync.RWMutex{},
		site:        make(map[string]*perSite),
		scrape:      &scraper.Scraper{},
		maxWorkers: semaphore.NewWeighted(100),
	}
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
	site.queue = append(site.queue, &siteURL)
	//site.queueSem.Signal()
	return
}

func (x *siteClient) newSite(siteURL string) *perSite {
	var err error
	out := &perSite{
		siteURL: siteURL,
		client: &fasthttp.Client{
			Name:            "chicory-scraper",
			MaxConnsPerHost: 100,
		},
		ctx:      context.Background(),
		limiter:  rate.NewLimiter(rate.Limit(1), 1),
		queue:    make([]*string, 0),
		queueSem: sync.NewCond(&sync.RWMutex{}),
		parent:   x,
	}
	go out.queueHandler()

	cacheCfg := ristretto.Config{
		NumCounters: 1e5,     // number of keys to track frequency of (100k).
		MaxCost:     1 << 16, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	}

	if out.cache, err = ristretto.NewCache(&cacheCfg); err != nil {
		log.Fatal(err)
	}
	return out
}

func (x *perSite) queueHandler() {
	for {
		// pull URL off the queue
		if len(x.queue) == 0 {
			// TODO sync with sender
			time.Sleep(time.Second)
			continue
		}
		siteURL := x.queue[0]
		x.queue[0] = nil
		x.queue = x.queue[1:]

		// rate limiter for the domain to avoid overloading website
		if err := x.limiter.Wait(x.ctx); err != nil {
			log.Warn(err)
		}

		dst := make([]byte, 128*1024)
		x.parent.maxWorkers.Acquire(x.ctx, 1)
		statusCode, body, err := x.client.GetTimeout(dst, *siteURL, time.Second*60)
		x.parent.maxWorkers.Release(1)

		switch err {
		case nil:
			// successfully communicated with a domain name
			switch statusCode {
			case fasthttp.StatusOK:
				recipe, found := x.parent.scrape.ScrapeRecipe(*siteURL, body)
				if !found {
					err = NotFoundError{"No Recipe Found"}
					return
				}
				x.parent.ReplyChan <- recipe
				err = nil
			default:
				recipe := &scraper.RecipeObject{
					SiteURL: *siteURL,
					StatusCode: statusCode,
					Error: "HTTP error",
				}
				x.parent.ReplyChan <- recipe
				err = fmt.Errorf("HTTP status code %d", statusCode)
				return
			}
		default:
			recipe := &scraper.RecipeObject{
				SiteURL: *siteURL,
				StatusCode: http.StatusServiceUnavailable,
				Error: err.Error(),
			}
			x.parent.ReplyChan <- recipe
		}
	}
}
