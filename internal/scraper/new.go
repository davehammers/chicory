// This package is the HTTP client that will connect with a website and extract the recipe information

package scraper

import (
	"context"
	"sync"

	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

type Scraper struct {
	cache         *ristretto.Cache
	badDomainMap  map[string]int
	badDomainLock *sync.RWMutex
	ctx           context.Context
	maxWorkers    *semaphore.Weighted
}

// NewClient - allocate a *Scraper
func New() *Scraper {
	var err error

	c := Scraper{
		ctx:           context.Background(),
		maxWorkers:    semaphore.NewWeighted(1000),
		badDomainMap:  make(map[string]int),
		badDomainLock: &sync.RWMutex{},
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
