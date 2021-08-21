package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
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
	var body []byte
	// first, look for cached recipe
	if r, ok := x.scrape.CachedRecipe(siteUrl); ok {
		recipe = &r
		return
	}
	// call rate limiter for the domain to avoid overloading website
	x.waitForDomain(siteUrl)

	// limit the number of concurrent transactions to avoid overloading the network
	req, err := http.NewRequest(http.MethodGet, siteUrl, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent","Onetsp-RecipeParser/0.1 (+https://github.com/onetsp/RecipeParser)" )
	x.maxWorkers.Acquire(x.ctx, 1)
	resp, err := x.client.Do(req)
	x.maxWorkers.Release(1)

	// Continue by getting recipe web page

	switch err {
	case nil:
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = NotFoundError{"No data returned"}
			return
		}
		resp.Body.Close()
		// successfully communicated with a domain name
		switch resp.StatusCode {
		case http.StatusOK:
			found := false
			recipe, found = x.scrape.ScrapeRecipe(siteUrl, body)
			if !found {
				err = NotFoundError{"No Recipe Found"}
				return
			}
			err = nil
		default:
			err = fmt.Errorf("HTTP status code %d", resp.StatusCode)
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

