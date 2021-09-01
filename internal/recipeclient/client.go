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

	"github.com/google/brotli/go/cbrotli"
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
func (x *RecipeClient) GetRecipe(sourceURL string) (recipe *scraper.RecipeObject, err error) {
	var body []byte
	// first, look for cached recipe
	if r, ok := x.scrape.CachedRecipe(sourceURL); ok {
		recipe = &r
		return
	}
	// call rate limiter for the domain to avoid overloading website
	x.waitForDomain(sourceURL)

	// limit the number of concurrent transactions to avoid overloading the network
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", RandomUserAgent())
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	//req.Header.Set("accept","*/*")
	req.Header.Set("referer", sourceURL)
	if u, err := url.Parse(sourceURL); err != nil {
		req.Header.Set("host", u.Host)
	}

	x.maxWorkers.Acquire(x.ctx, 1)
	resp, err := x.client.Do(req)
	x.maxWorkers.Release(1)

	// Continue by getting recipe web page

	switch err {
	case nil:
		defer resp.Body.Close()
		// gzip is handled by the lower transport automatically
		switch resp.Header.Get("Content-Encoding") {
		case "br":
			// some web sites compress data even if not requested
			reader := cbrotli.NewReader(resp.Body)
			defer reader.Close()
			body, err = ioutil.ReadAll(reader)
		default:
			body, err = ioutil.ReadAll(resp.Body)
		}

		if err != nil {
			err = NotFoundError{"No data returned"}
			return
		}
		// successfully communicated with a domain name
		switch resp.StatusCode {
		case http.StatusOK:
			found := false
			recipe, found = x.scrape.ScrapeRecipe(sourceURL, body)
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

func (x *RecipeClient) waitForDomain(sourceURL string) {
	u, err := url.Parse(sourceURL)
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
