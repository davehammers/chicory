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
	"strings"
	"sync"
	"testing"

	"scraper/internal/recipeclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
	log "github.com/sirupsen/logrus"

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
	go func() {
		for recipe := range siteClient.ReplyChan {
			wg.Done()
			switch recipe.StatusCode {
			case http.StatusOK:
				b, err := JSONMarshal(recipe)
				if err == nil {
					log.Println("Found", recipe.SiteURL)
					log.Println(string(b))
				}
			default:
				log.Println(recipe.StatusCode, recipe.SiteURL, recipe.Error)
			}
		}
	}()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		siteURL := sc.Text() // GET the line string
		siteURL = strings.TrimRight(siteURL, "\n")
		wg.Add(1)
		err := siteClient.SiteGetRecipe(siteURL)
		if err != nil {
			log.Println("URL Error", siteURL, err)
		}
	}
	wg.Wait()
}
