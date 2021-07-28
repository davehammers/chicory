package recipeclient

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"strings"
)

const (
	RecipeType = "Recipe"
)

type NotFoundError struct {
	message string
}

func (x NotFoundError) Error() string {
	return x.message
}

type RecipeSchema1 struct {
	RecipeIngredient   []string    `json:"recipeIngredient"`
}

type RecipeSchema2 struct {
	Context string `json:"@context"`
	Graph   []struct {
		RecipeIngredient   []string `json:"recipeIngredient,omitempty"`
	} `json:"@graph"`
}

// GetRecipe - extract any recipes from the URL site provided. returns interface
func (x *RecipeClient) GetRecipe(siteUrl string) (recipe interface{}, err error) {
	// first, look for cached recipe
	/*
	cRecipe, err := x.cachedRecipe(siteUrl)
	if err == nil {
		// found
		recipe = cRecipe
		return
	}
	 */

	// Continue by getting recipe web page
	// format a GET request
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(siteUrl)
	err = x.client.DoRedirects(req, resp, 5)
	if err != nil {
		return
	}

	recipe, err = x.parseForJSONRecipe(siteUrl, resp.Body())
	return
}

// parseForJSONRecipe parse through the HTML looking for schema.org json-ld format
func (x *RecipeClient) parseForJSONRecipe(siteUrl string, body []byte) (recipe interface{}, err error) {
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err = nil
			return
		case html.TextToken:
			text := tokenizer.Text()
			if !strings.Contains(string(text), RecipeType) {
				continue
			}
			r1 := RecipeSchema1{}
			err = json.Unmarshal(text, &r1)
			if err == nil {
				//fmt.Println(string(text))
				if len(r1.RecipeIngredient) > 0 {
					recipe = r1
					x.addRecipeToCache(siteUrl, &recipe)
					return
				}
			}
			r2 := RecipeSchema2{}
			err = json.Unmarshal(text, &r2)
			if err == nil {
				for _, entry := range r2.Graph{
					if len(entry.RecipeIngredient) > 0 {
						recipe = r2
						x.addRecipeToCache(siteUrl, &recipe)
						return
					}
				}
			}
		}
	}
}

func (x *RecipeClient) cachedRecipe(siteUrl string) (recipe interface{}, err error) {
	// cache lookup to avoid HTTP transactions
	recipe, found := x.cache.Get(siteUrl)
	if !found {
		err = NotFoundError{"No cache recipe"}
		return
	}
	return
}

func (x *RecipeClient) addRecipeToCache(siteUrl string, recipe interface{}) (err error) {
	x.cache.Set(siteUrl, recipe, 1)
	x.cache.Wait()
	return
}
