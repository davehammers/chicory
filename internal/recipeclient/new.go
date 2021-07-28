// This package is the HTTP client that will connect with a website and extract the recipe information

package recipeclient

import (
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type RecipeClient struct {
	client *fasthttp.Client
	cache  *ristretto.Cache
}

// NewClient - allocate a *RecipeClient
func New() *RecipeClient {
	var err error

	c := RecipeClient{}
	c.client = &fasthttp.Client{
		Name: "chicory-scraper",
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
