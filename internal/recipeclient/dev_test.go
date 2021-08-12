// +build dev

package recipeclient_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"scraper/internal/recipeclient"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

func TestRecipeClient_GetRecipe(t *testing.T) {
	x := recipeclient.New()
	f, err := os.Open("testdata/recipeURLs")
	require.Nil(t, err)
	defer f.Close()

	ctx := context.Background()
	sem := semaphore.NewWeighted(50000)
	wg := sync.WaitGroup{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		siteURL := sc.Text() // GET the line string
		siteURL = strings.TrimRight(siteURL, "\n")
		wg.Add(1)
		sem.Acquire(ctx, 1)
		go func(siteURL string) {
			defer wg.Done()
			defer sem.Release(1)
			recipe, err := x.GetRecipe(siteURL)
			if err == nil {
				b, err := JSONMarshal(recipe)
				if err == nil {
					fmt.Println("Found", siteURL)
					fmt.Println(string(b))
				}
			} else {
				switch err.(type) {
				case recipeclient.NotFoundError:
					fmt.Println("NotFound", siteURL)
				case recipeclient.TimeOutError:
					fmt.Println("Timeout", siteURL)
				default:
					fmt.Println("Unknown", siteURL, err)
				}
			}
		}(siteURL)
	}
	assert.Nil(t, sc.Err())
	wg.Wait()
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// HTTP summary
type bySite map[string]int
type byCode map[int]int
type byScraper map[string]int

type siteCode map[string]byCode
type codeSite map[int]bySite

// by scraper
type siteScraper map[string]byScraper
type scraperSite map[string]bySite

func TestSiteClient_SiteGetRecipe(t *testing.T) {
	siteClient := recipeclient.NewSiteClient()
	f, err := os.Open("testdata/recipeURLs")
	require.Nil(t, err)
	defer f.Close()
	wg := sync.WaitGroup{}
	sumCodeSite := make(codeSite)
	sumSiteCode := make(siteCode)
	sumSiteScraper := make(siteScraper)
	sumScraperSite := make(scraperSite)

	wg.Add(1)
	go func() {
		activity := false
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ticker.C:
				if !activity {
					wg.Done()
					return
				}
				activity = false
			case recipe := <-siteClient.ReplyChan:
				// count by HTTP code
				activity = true
				u, _ := url.Parse(recipe.SiteURL)
				// count statusCode/Site
				if _, ok := sumCodeSite[recipe.StatusCode]; !ok {
					sumCodeSite[recipe.StatusCode] = make(bySite)
				}
				sumCodeSite[recipe.StatusCode][u.Host]++

				// count site/StatusCode
				if _, ok := sumSiteCode[u.Host]; !ok {
					sumSiteCode[u.Host] = make(byCode)
				}
				sumSiteCode[u.Host][recipe.StatusCode]++

				switch recipe.StatusCode {
				case http.StatusOK:
					scraper := recipe.Scraper[0]
					// count sites per scraper
					if _, ok := sumScraperSite[scraper]; !ok {
						sumScraperSite[scraper] = make(bySite)
					}
					sumScraperSite[scraper][u.Host]++
					// count scrapers per site
					if _, ok := sumSiteScraper[u.Host]; !ok {
						sumSiteScraper[u.Host] = make(byScraper)
					}
					sumSiteScraper[u.Host][scraper]++

					b, err := JSONMarshal(recipe)
					if err == nil {
						log.Println(recipe.StatusCode, recipe.SiteURL)
						fmt.Println(string(b))
					}
				default:
					log.Println(recipe.StatusCode, recipe.SiteURL, recipe.Error)
				}
			}
		}
	}()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		siteURL := sc.Text() // GET the line string
		siteURL = strings.TrimRight(siteURL, "\n")
		if siteURL == "" {
			continue
		}
		err := siteClient.SiteGetRecipe(siteURL)
		if err != nil {
			log.Println("URL Error", siteURL, err)
		}
	}
	wg.Wait()
	fmt.Println(strings.Repeat("-", 80))
	printHttpSummary(sumCodeSite, sumSiteCode)
	printSiteSummary(sumSiteScraper, sumScraperSite)
	printScraperSummary(sumSiteScraper, sumScraperSite)
}
func printHttpSummary(sumCodeSite codeSite, sumSiteCode siteCode) {
	type sortType struct {
		code  int
		count int
	}
	type sortStringType struct {
		text  string
		count int
	}
	typeList := make([]sortType, 0)
	// HTTP Summary
	for k, v := range sumCodeSite {
		count := 0
		// add up all of the sites
		for _, v1 := range v {
			count += v1
		}
		typeList = append(typeList, sortType{k, count})
	}
	sort.Slice(typeList, func(i, j int) bool { return typeList[i].count > typeList[j].count })

	// print detail
	fmt.Println("")
	fmt.Println("HTTP Detail")
	for _, t := range typeList {
		fmt.Printf("Count:%6d, HTTP: %d\n", t.count, t.code)
		detailList := make([]sortStringType, 0)
		for k, v := range sumCodeSite[t.code] {
			detailList = append(detailList, sortStringType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Site: %s\n", d.count, d.text)
		}
	}
	total := 0
	fmt.Println("")
	fmt.Println("HTTP Summary")
	for _, t := range typeList {
		total += t.count
		fmt.Printf("Count:%6d, HTTP: %d\n", t.count, t.code)
	}
	fmt.Printf("Total:%6d, HTTP\n", total)

}

func printSiteSummary(sumSiteScraper siteScraper, sumScraperSite scraperSite) {
	type sortType struct {
		text  string
		count int
	}
	textList := make([]sortType, 0)
	total := 0
	for k, v := range sumSiteScraper {
		count := 0
		for _, v1 := range v {
			count += v1
		}
		textList = append(textList, sortType{k, count})
	}
	sort.Slice(textList, func(i, j int) bool { return textList[i].count > textList[j].count })

	fmt.Println("")
	fmt.Println("Site Detail")
	// print detail
	for _, t := range textList {
		fmt.Printf("Count:%6d, Site: %s\n", t.count, t.text)
		detailList := make([]sortType, 0)
		for k, v := range sumSiteScraper[t.text] {
			detailList = append(detailList, sortType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Scraper: %s\n", d.count, d.text)
		}
	}
	total = 0
	fmt.Println("")
	fmt.Println("Site Summary")
	for _, t := range textList {
		total += t.count
		fmt.Printf("Count:%6d, Site: %s\n", t.count, t.text)
	}
	fmt.Printf("Total:%6d, Site\n", total)
}
func printScraperSummary(sumSiteScraper siteScraper, sumScraperSite scraperSite) {
	type sortType struct {
		text  string
		count int
	}
	textList := make([]sortType, 0)
	total := 0
	for k, v := range sumScraperSite {
		count := 0
		for _, v1 := range v {
			count += v1
		}
		textList = append(textList, sortType{k, count})
	}
	sort.Slice(textList, func(i, j int) bool { return textList[i].count > textList[j].count })

	fmt.Println("")
	fmt.Println("Scraper Detail")
	// print detail
	for _, t := range textList {
		fmt.Printf("Count:%6d, Scraper: %s\n", t.count, t.text)
		detailList := make([]sortType, 0)
		for k, v := range sumScraperSite[t.text] {
			detailList = append(detailList, sortType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Site: %s\n", d.count, d.text)
		}
	}
	// calculate total
	total = 0
	for _, t := range textList {
		total += t.count
	}
	fmt.Println("")
	fmt.Println("Scraper Summary")
	for _, t := range textList {
		fmt.Printf("Count:%6d, %.1f%% Scraper: %s\n", t.count, float32(t.count)/float32(total) * 100.0, t.text)
	}
	fmt.Printf("Total:%6d, Scraper\n", total)
}
