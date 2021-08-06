package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
	"net/url"
	"scraper/internal/scraper"
	"time"
)

const (
	ReadTimeout = time.Second * 60
	RecipeType  = "Recipe"
	LdType      = "@type"
)

type NotFoundError struct {
	message string
}

func (x NotFoundError) Error() string {
	return x.message
}

type TimeOutError struct {
	message string
}

func (x TimeOutError) Error() string {
	return x.message
}

// GetRecipe - extract any recipes from the URL site provided. returns interface
func (x *RecipeClient) GetRecipe(siteUrl string) (recipe *scraper.RecipeObject, err error) {
	// first, look for cached recipe
	if r, ok := x.scrape.CachedRecipe(siteUrl); ok {
		recipe = &r
		return
	}
	// call rate limiter for the domain to avoid overloading website
	x.waitForDomain(siteUrl)

	// limit the number of concurrent transactions to avoid overloading the network
	x.maxWorkers.Acquire(x.ctx, 1)
	defer x.maxWorkers.Release(1)

	// Continue by getting recipe web page
	// format a GET request

	dst := make([]byte,128*1024)
	statusCode, body, err := x.client.GetTimeout(dst, siteUrl, time.Second * 60)

	//req := fasthttp.AcquireRequest()
	//resp := fasthttp.AcquireResponse()
	//req.SetRequestURI(siteUrl)
	//err = x.client.DoTimeout(req, resp, time.Second * 60)
	switch err {
	case nil:
		// successfully communicated with a domain name
		switch statusCode {
		case fasthttp.StatusOK:
			found := false
			recipe, found = x.scrape.ScrapeRecipe(siteUrl, body)
			if !found {
				err = NotFoundError{"No Recipe Found"}
				return
			}
			err = nil
		default:
			err = fmt.Errorf("HTTP status code %d", statusCode)
			return
		}
	}
	return
}

func (x *RecipeClient) waitForDomain(siteUrl string) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return
	}
	x.domainRateLock.Lock()
	limiter, ok := x.domainRateMap[u.Host]
	if !ok {
		// domain does not have a rate limiter yet
		// set one up
		limiter = rate.NewLimiter(rate.Limit(1), 1)
		x.domainRateMap[u.Host] = limiter
	}
	x.domainRateLock.Unlock()
	if err = limiter.Wait(x.ctx); err != nil {
		log.Warn(err)
	}
}

