// This package is the HTTP client that will connect with a website and extract the recipe information

package recipeclient

import (
	"context"
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/semaphore"
	"scraper/internal/scraper"
	"sync"
	"time"
)

type RecipeClient struct {
	client *fasthttp.Client
	cache  *ristretto.Cache
	badDomainMap map[string]int
	badDomainLock *sync.RWMutex
	ctx context.Context
	maxWorkers *semaphore.Weighted
	scrape *scraper.Scraper
}

// NewClient - allocate a *RecipeClient
func New() *RecipeClient {
	var err error

	c := RecipeClient{
		ctx: context.Background(),
		maxWorkers: semaphore.NewWeighted(200),
		badDomainMap: make(map[string]int),
		badDomainLock: &sync.RWMutex{},
		scrape: &scraper.Scraper{},
	}
	c.client = &fasthttp.Client{
		Name: "chicory-scraper",
		MaxConnsPerHost: 100,
		//MaxConnDuration: time.Second * 60,
		ReadBufferSize: 8*1024,
		ReadTimeout: time.Second * 60,
	}
	cacheCfg := ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	}
	if c.cache, err = ristretto.NewCache(&cacheCfg); err != nil {
		log.Fatal(err)
	}
	return &c
}
