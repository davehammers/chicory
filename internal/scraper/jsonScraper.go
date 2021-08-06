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
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return false
		case html.TextToken:
			text := tokenizer.Text()
			//if !strings.Contains(string(text), LdType) {
			//continue
			//}
			switch {
			case x.jsonSchema1(siteUrl, text, recipe):
				return true
			case x.jsonSchema2(siteUrl, text, recipe):
				return true
			case x.jsonSchema3(siteUrl, text, recipe):
				return true
			case x.jsonSchema4(siteUrl, text, recipe):
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

// jsonSchema1 parse json schema 1
func (x *Scraper) jsonSchema1(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema1{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		if len(r.RecipeIngredient) > 0 {
			recipe.Type = JSON1RecipeType
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

// jsonSchema2 parse json schema 2
func (x *Scraper) jsonSchema2(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.Graph {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Type = JSON2RecipeType
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

// jsonSchema3 parse json schema 3
func (x *Scraper) jsonSchema3(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema3{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Type = JSON3RecipeType
				x.jsonAppend(recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}

// jsonSchema3 parse json schema 3
func (x *Scraper) jsonAppend(recipe *RecipeObject, list []string) {
	for _, text := range list{
		if text == "" {
			continue
		}
		text = strings.TrimSpace(text)
		recipe.RecipeIngredient =  append(recipe.RecipeIngredient, text)
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

// jsonSchema4 parse json schema 4
// https://www.yummly.com/recipe/Roasted-garlic-caesar-dipping-sauce-297499
func (x *Scraper) jsonSchema4(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema4{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.ItemListElement {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Type = JSON4RecipeType
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
func (x *Scraper) jsonSchemaRemoveHTML(siteUrl string, body []byte, recipe *RecipeObject) (ok bool) {
	textOut := make([]string, 0)
	// is this a schema.org JSON string
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
	log.Info(string(text))
	switch {
	case x.jsonSchema1(siteUrl, text, recipe):
		return true
	case x.jsonSchema2(siteUrl, text, recipe):
		return true
	case x.jsonSchema3(siteUrl, text, recipe):
		return true
	case x.jsonSchema4(siteUrl, text, recipe):
		return true
	}
	log.Error(siteUrl, "body contains schema.org/recipeIngredient json but did not parse")
	return false
}
