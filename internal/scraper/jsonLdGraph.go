package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"strings"
)

const (
	RecipeType = "Recipe"
	LdType     = "@type"
)

// jsonParser tries to extract recipe in JSON-LD format
func (x *Scraper) jsonParser(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	insideScript := false
	if strings.Contains(string(body), `"@type":"Recipe"`) ||
		strings.Contains(string(body), `"@type": "Recipe"`) {
		recipe.Attributes = append(recipe.Attributes, "@type:Recipe")
	}
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "script":
				if strings.Contains(string(body), "schema.org") {
					insideScript = true
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "script":
				insideScript = false
			}

		case html.TextToken:
			if !insideScript {
				break
			}
			text := tokenizer.Text()
			switch {
			case x.graph_schemaOrgJSON(siteUrl, text, recipe):
				return true
			case x.schemaOrg_RecipeJSON(siteUrl, text, recipe):
				return true
			case x.schemaOrg_List(siteUrl, text, recipe):
				return true
			case x.schemaOrg_ItemListJSON(siteUrl, text, recipe):
				return true
			case x.jsonSchemaRemoveHTML(siteUrl, text, recipe):
				return true
			default:
				continue
			}
		}
	}
	return false
}

// flat schema.org recipe
type RecipeSchema1 struct {
	RecipeIngredient []string `json:"recipeIngredient"`
}

// schemaOrg_RecipeJSON parse json schema 1
// http://30pepperstreet.com/recipe/endive-salad/
func (x *Scraper) schemaOrg_RecipeJSON(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema1{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		if len(r.RecipeIngredient) > 0 {
			recipe.Scraper = append(recipe.Scraper, "schemaOrgRecipe")
			x.jsonAppend(recipe, r.RecipeIngredient)
			return true
		}
	}
	return false
}

// nexted in Graph data
type RecipeSchema2 struct {
	Context string `json:"@context"`
	Graph   []struct {
		RecipeIngredient []string `json:"recipeIngredient,omitempty"`
	} `json:"@graph"`
}

// graph_schemaOrgJSON parse json schema 2
// http://ahealthylifeforme.com/25-minute-garlic-mashed-potatoes
func (x *Scraper) graph_schemaOrgJSON(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.Graph {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "@graph-schemaOrgRecipe")
				x.jsonAppend(recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}

// top level list
type RecipeSchema3 []struct {
	Context          string   `json:"@context"`
	Type             string   `json:"@type"`
	RecipeIngredient []string `json:"recipeIngredient,omitempty"`
}

// schemaOrg_List parse json schema 3
// http://allrecipes.com/recipe/12646/cheese-and-garden-vegetable-pie/
func (x *Scraper) schemaOrg_List(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema3{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "list schemaOrg")
				x.jsonAppend(recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}

// schemaOrg_List parse json schema 3
// http://ahealthylifeforme.com/25-minute-garlic-mashed-potatoes
func (x *Scraper) jsonAppend(recipe *RecipeObject, list []string) {
	for _, text := range list {
		if text == "" {
			continue
		}
		text = strings.TrimSpace(text)
		recipe.RecipeIngredient = append(recipe.RecipeIngredient, text)
	}
}

type RecipeSchema4 struct {
	Context         string `json:"@context"`
	Type            string `json:"@type"`
	ItemListElement []struct {
		RecipeIngredient []string `json:"recipeIngredient"`
	} `json:"itemListElement"`
	NumberOfItems int    `json:"numberOfItems"`
	ItemListOrder string `json:"itemListOrder"`
}

// schemaOrg_ItemListJSON parse json schema 4
// https://www.yummly.com/recipe/Roasted-garlic-caesar-dipping-sauce-297499
func (x *Scraper) schemaOrg_ItemListJSON(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema4{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.ItemListElement {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "schemaOrgItemList")
				recipe.RecipeIngredient = entry.RecipeIngredient
				return true
			}
		}
	}
	return false
}

// jsonSchemaRemoveHTML parse json schema RemoveHTML
// this parser assumes that there is HTML mixed in with the JSON
// It tries to remove all of the mixed in HTML then reprocess the JSON
// https://mealpreponfleek.com/low-carb-hamburger-helper/
func (x *Scraper) jsonSchemaRemoveHTML(siteUrl string, body []byte, recipe *RecipeObject) (ok bool) {
	textOut := make([]string, 0)
	// is this a schema.org JSON string
	if strings.Contains(string(body), "ld+json") {
		recipe.Attributes = append(recipe.Attributes, "ld+json")
	}
	if !strings.Contains(string(body), "recipeIngredient") {
		return false
	}

	// Lets try to fix the JSON by parsing out all of the html
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	done := false
	for !done {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			done = true
			break
		case html.TextToken:
			text := string(tokenizer.Text())
			text = strings.ReplaceAll(text, "\n", " ")
			text = strings.ReplaceAll(text, "\r", " ")
			text = strings.TrimSpace(text)
			if len(text) == 0 {
				continue
			}
			textOut = append(textOut, text)
		}
	}
	// now try the json parsers again
	text := []byte(strings.Join(textOut, " "))
	switch {
	case x.schemaOrg_RecipeJSON(siteUrl, text, recipe):
		recipe.Attributes = append(recipe.Attributes, "JSON repaired")
		return true
	case x.graph_schemaOrgJSON(siteUrl, text, recipe):
		recipe.Attributes = append(recipe.Attributes, "JSON repaired")
		return true
	case x.schemaOrg_List(siteUrl, text, recipe):
		recipe.Attributes = append(recipe.Attributes, "JSON repaired")
		return true
	case x.schemaOrg_ItemListJSON(siteUrl, text, recipe):
		recipe.Attributes = append(recipe.Attributes, "JSON repaired")
		return true
	}
	log.Error(siteUrl, "body contains schema.org/recipeIngredient json but did not parse")
	recipe.Attributes = append(recipe.Attributes, "No JSON Parser")
	return false
}
