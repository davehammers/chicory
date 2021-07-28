package recipeclient

// contains definitions and functions for accessing and parsing recipies from URLs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
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

type SchemaRecipe struct {
	Graph []struct {
		Type            string `json:"@type"`
		Text            string `json:"text"`
		SuggestedAnswer []struct {
			Type string `json:"@type"`
			Text string `json:"text"`
		} `json:"suggestedAnswer"`
	} `json:"@graph"`
	Context   string `json:"@context"`
	URL       string `json:"url"`
	Publisher struct {
		Type string `json:"@type"`
		Name string `json:"name"`
		Logo struct {
			Type   string `json:"@type"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"logo"`
	} `json:"publisher"`
	Type   string `json:"@type"`
	Author struct {
		Name string `json:"name"`
		Type string `json:"@type"`
	} `json:"author"`
	DatePublished time.Time `json:"datePublished"`
	Headline      string    `json:"headline"`
	/*
	Image         *struct {
		Type      string `json:"@type"`
		Height    int    `json:"height"`
		Thumbnail string `json:"thumbnail"`
		URL       string `json:"url"`
		Width     int    `json:"width"`
	} `json:"image"`
	 */
	MainEntityOfPage struct {
		ID   string `json:"@id"`
		Type string `json:"@type"`
	} `json:"mainEntityOfPage"`
	ThumbnailURL        string    `json:"thumbnailUrl"`
	DateModified        time.Time `json:"dateModified"`
	IsAccessibleForFree string    `json:"isAccessibleForFree"`
	HasPart             []struct {
		Type                string `json:"@type"`
		IsAccessibleForFree string `json:"isAccessibleForFree"`
		CSSSelector         string `json:"cssSelector"`
	} `json:"hasPart"`
	Name               string   `json:"name"`
	PrepTime           string   `json:"prepTime"`
	CookTime           string   `json:"cookTime"`
	TotalTime          string   `json:"totalTime"`
	RecipeIngredient   []string `json:"recipeIngredient"`
	RecipeInstructions interface{}   `json:"recipeInstructions"`
	Video              struct {
		Type         string    `json:"@type"`
		ContentURL   string    `json:"contentUrl"`
		Description  string    `json:"description"`
		Duration     string    `json:"duration"`
		EmbedURL     string    `json:"embedUrl"`
		Name         string    `json:"name"`
		ThumbnailURL string    `json:"thumbnailUrl"`
		UploadDate   time.Time `json:"uploadDate"`
	} `json:"video"`
	RecipeCuisine   []string `json:"recipeCuisine"`
	/*
	AggregateRating struct {
		Context     string  `json:"@context"`
		Type        string  `json:"@type"`
		RatingValue fload64 `json:"ratingValue"`
		ReviewCount int     `json:"reviewCount"`
		WorstRating int     `json:"worstRating"`
		BestRating  int     `json:"bestRating"`
	} `json:"aggregateRating"`
	*/
	Review []struct {
		Context       string `json:"@context"`
		Type          string `json:"@type"`
		DatePublished string `json:"datePublished"`
		Author        struct {
			Context    string `json:"@context"`
			Type       string `json:"@type"`
			Name       string `json:"name"`
			GivenName  string `json:"givenName"`
			FamilyName string `json:"familyName"`
		} `json:"author,omitempty"`
		ReviewBody   string `json:"reviewBody"`
		ReviewRating struct {
			Context     string `json:"@context"`
			Type        string `json:"@type"`
			RatingValue int    `json:"ratingValue"`
			WorstRating int    `json:"worstRating"`
			BestRating  int    `json:"bestRating"`
		} `json:"reviewRating"`
	} `json:"review"`
	RecipeCategory interface{} `json:"recipeCategory"`
	RecipeYield    string   `json:"recipeYield"`
	Description    string   `json:"description"`
	Keywords       string   `json:"keywords"`
	amount string
	description string
}

type itemType struct {
	amount      string
	description string
}

// GetRecipe - extract any recipies from the URL site provided. returns []SchemaRecipe
func (x *RecipeClient) GetRecipe(siteUrl string) (recipe SchemaRecipe, err error) {
	// first, look for cached recipe
	cRecepe, err := x.cachedRecipe(siteUrl)
	if  err == nil {
		// found
		recipe = *cRecepe
		return
	}

	// Continue by getting recipe web page
	// format a GET request
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	req.SetRequestURI(siteUrl)
	err = x.client.DoRedirects(req, resp, 5)
	if err != nil {
		return
	}

	// parse the HTML body
	doc, err := html.Parse(bytes.NewReader(resp.Body()))

	// Next try looking for JSON recipes
	err = x.jsonRecipe(siteUrl,doc, &recipe)
	switch err.(type) {
	case nil:
		// ingredient cleanup
		for idx,_ := range recipe.RecipeIngredient{
			recipe.RecipeIngredient[idx] = strings.ReplaceAll(recipe.RecipeIngredient[idx], "<p>","")
			recipe.RecipeIngredient[idx] = strings.ReplaceAll(recipe.RecipeIngredient[idx], "</p>","")
		}
		return
	}

	// Next scraper pulls recipes from the HTML
	err = x.htmlRecipe(doc, &recipe)
	switch err.(type) {
	case nil:
		// found
		return
	}

	return
}

func (x *RecipeClient)cachedRecipe(siteUrl string) (recipe *SchemaRecipe, err error) {
	// cache lookup to avoid HTTP transactions
	rIntf, found := x.cache.Get(siteUrl)
	if !found {
		err = NotFoundError{"No cache recipe"}
		return
	}
	r :=SchemaRecipe{}
	r = rIntf.(SchemaRecipe)
	recipe = &r
	return
}

func (x *RecipeClient)addRecipeToCache(siteUrl string, recipe *SchemaRecipe) (err error) {
	x.cache.Set(siteUrl, *recipe, 1)
	x.cache.Wait()
	return
}

// jsonRecipe - local function to walk through the HTML nodes looking for recipes in JSON-LD format
func (x *RecipeClient)jsonRecipe(siteUrl string,n *html.Node, recipe *SchemaRecipe) (err error) {
	for _, attribute := range n.Attr {
		switch attribute.Val {
		case "json-ld", "application/ld+json":
			for p := n.FirstChild; p != nil; p = p.FirstChild {
				if err = json.Unmarshal([]byte(p.Data), recipe);err == nil {
					// add recipe to cache
					x.addRecipeToCache(siteUrl,recipe)
				}
				return
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err = x.jsonRecipe(siteUrl,c, recipe)
		if err == nil {
			return
		}
	}
	return NotFoundError{"No recipe found"}
}

// htmoRecipe - local function to walk through the HTML nodes looking for recipes in JSON-LD format
func (x *RecipeClient)htmlRecipe(n *html.Node, recipe *SchemaRecipe) (err error) {
	for _, attribute := range n.Attr {
		switch attribute.Val {
		case "ingredient-amount":
			recipe.amount = n.FirstChild.Data
			recipe.amount = strings.ReplaceAll(recipe.amount, "\n", "")
			recipe.amount = strings.ReplaceAll(recipe.amount, "\t", "")
		case "ingredient-description":
			description := ""
			for p := n.FirstChild; p != nil; p = p.FirstChild {
				if p.Type == 1 {
					description += p.Data
				}
			}
			recipe.description = description
			recipe.RecipeIngredient = append(recipe.RecipeIngredient, fmt.Sprintf("%s %s", recipe.amount, recipe.description))
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err = x.htmlRecipe(c, recipe)
		if err == nil {
			return
		}
	}
	if len(recipe.RecipeIngredient) > 0 {
		return
	}
	return NotFoundError{"No recipe found"}
}
