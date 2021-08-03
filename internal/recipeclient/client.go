package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"fmt"
	"github.com/valyala/fasthttp"
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
	if x.domainIsBad(siteUrl) {
		err = TimeOutError{"Domain name attempts exceeded"}
		return
	}

	// limit the number of concurrent transactions to avoid overloading the network
	x.maxWorkers.Acquire(x.ctx, 1)
	defer x.maxWorkers.Release(1)

	// Continue by getting recipe web page
	// format a GET request

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(siteUrl)

	/*
		for retry := 0; retry < 2; retry++ {
			if err == nil {
				break
			}
		}
	*/
	//err = x.client.DoRedirects(req, resp, 3)
	err = x.client.DoRedirects(req, resp, 100)
	switch err {
	case nil:
		// successfully communicated with a domain name
		x.domainIsGood(siteUrl)

		switch resp.StatusCode() {
		case fasthttp.StatusOK:
			found := false
			recipe, found = x.scrape.ScrapeRecipe(siteUrl, resp.Body())
			if !found {
				err = NotFoundError{"No Recipe Found"}
				return
			}
			err = nil
		case fasthttp.StatusMovedPermanently, fasthttp.StatusFound, fasthttp.StatusTemporaryRedirect, fasthttp.StatusPermanentRedirect:
			err = NotFoundError{"Recipe site has moved or is no longer available"}
			return
		default:
			err = fmt.Errorf("HTTP status code %d", resp.StatusCode())
			return
		}
	case fasthttp.ErrTimeout:
		err = TimeOutError{err.Error()}
		return
	case fasthttp.ErrNoFreeConns:
		err = TimeOutError{err.Error()}
		return
	}
	return
}

func (x *RecipeClient) domainIsBad(siteUrl string) (bad bool) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return false
	}
	x.badDomainLock.RLock()
	cnt, ok := x.badDomainMap[u.Host]
	x.badDomainLock.RUnlock()
	if !ok {
		return false
	}
	if cnt > 10 {
		return true
	}
	return false
}

// domainIsGood domain name worked for http
func (x *RecipeClient) domainIsGood(siteUrl string) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return
	}
	x.badDomainLock.Lock()
	delete(x.badDomainMap, u.Host)
	x.badDomainLock.Unlock()
	return
}

func (x *RecipeClient) addBadDomain(siteUrl string) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return
	}
	x.badDomainLock.Lock()
	defer x.badDomainLock.Unlock()
	cnt, bad := x.badDomainMap[u.Host]
	if !bad {
		x.badDomainMap[u.Host] = 0
	}
	cnt++
	x.badDomainMap[u.Host]++
}
