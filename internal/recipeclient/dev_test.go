// +build dev

package recipeclient_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func TestSiteClient_SiteGetRecipe(t *testing.T) {
	siteClient := recipeclient.NewSiteClient()
	f, err := os.Open("testdata/recipeURLs")
	require.Nil(t, err)
	defer f.Close()
	wg := sync.WaitGroup{}
	httpSummary := make(map[int]int)
	typeSummary := make(map[string]int)

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
				if _, ok := httpSummary[recipe.StatusCode]; !ok {
					httpSummary[recipe.StatusCode] = 0
				}
				httpSummary[recipe.StatusCode]++

				// count by parser type
				if recipe.Scraper != nil {
					parser := strings.Join(recipe.Scraper ,",")
					if len(recipe.Attributes) > 0 {
						parser += ": "
						parser += strings.Join(recipe.Attributes ,",")
					}
					if _, ok := typeSummary[parser]; !ok {
						typeSummary[parser] = 0
					}
					typeSummary[parser]++
				}

				switch recipe.StatusCode {
				case http.StatusOK:
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
	type sortHttp struct {
		code  int
		count int
	}
	httpList := make([]sortHttp, 0)
	for k, v := range httpSummary {
		httpList = append(httpList, sortHttp{k, v})
	}
	sort.Slice(httpList, func(i, j int) bool { return httpList[i].count > httpList[j].count })
	fmt.Println("HTTP Summary")
	for _, h := range httpList {
		fmt.Printf("Count:%6d, HTTP: %d, %s\n", h.count, h.code, http.StatusText(h.code))
	}

	type sortType struct {
		parserType string
		count      int
	}
	typeList := make([]sortType, 0)
	for k, v := range typeSummary {
		typeList = append(typeList, sortType{k, v})
	}
	sort.Slice(typeList, func(i, j int) bool { return typeList[i].count > typeList[j].count })
	fmt.Println("")
	fmt.Println("Parser Summary")
	for _, t := range typeList {
		fmt.Printf("Count:%6d, Parser: %s\n", t.count, t.parserType)
	}
}
