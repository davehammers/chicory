// This package is the HTTP client that will connect with a website and extract the recipe information

package recipeclient

import (
	"context"
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"net/http"
	"scraper/internal/scraper"
	"sync"
	"time"
)

type RecipeClient struct {
	client *http.Client
	cache  *ristretto.Cache
	domainRateMap map[string]*rate.Limiter
	domainRateLock *sync.RWMutex
	ctx context.Context
	maxWorkers *semaphore.Weighted
	scrape *scraper.Scraper
}

// NewClient - allocate a *RecipeClient
func New() *RecipeClient {
	var err error

	c := RecipeClient{
		ctx: context.Background(),
		maxWorkers: semaphore.NewWeighted(5000),
		domainRateMap: make(map[string]*rate.Limiter),
		domainRateLock: &sync.RWMutex{},
		scrape: scraper.New(),
	}
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

	c.client = &http.Client{Transport: tr, Timeout: time.Second * 60}
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
