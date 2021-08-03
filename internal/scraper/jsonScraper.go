package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"encoding/json"
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
			case x.jsonSchema1(text, recipe):
				return true
			case x.jsonSchema2(text, recipe):
				return true
			case x.jsonSchema3(text, recipe):
				return true
			case x.jsonSchema4(text, recipe):
				return true
			default:
				continue
			}
		}
	}
	return false
}

type RecipeSchema1 struct {
	RecipeIngredient []string `json:"recipeIngredient"`
}

// jsonSchema1 parse json schema 1
func (x *Scraper) jsonSchema1(text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema1{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		if len(r.RecipeIngredient) > 0 {
			recipe.Type = JSON1RecipeType
			recipe.RecipeIngredient = r.RecipeIngredient
			return true
		}
	}
	return false
}

type RecipeSchema2 struct {
	Context string `json:"@context"`
	Graph   []struct {
		RecipeIngredient []string `json:"recipeIngredient,omitempty"`
	} `json:"@graph"`
}

// jsonSchema2 parse json schema 2
func (x *Scraper) jsonSchema2(text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.Graph {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Type = JSON2RecipeType
				recipe.RecipeIngredient = entry.RecipeIngredient
				return true
			}
		}
	}
	return false
}

type RecipeSchema3 []struct {
	Context          string   `json:"@context"`
	Type             string   `json:"@type"`
	RecipeIngredient []string `json:"recipeIngredient,omitempty"`
}

// jsonSchema3 parse json schema 3
func (x *Scraper) jsonSchema3(text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema3{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Type = JSON3RecipeType
				recipe.RecipeIngredient = entry.RecipeIngredient
				return true
			}
		}
	}
	return false
}

// jsonSchema4 parse json schema 3
// this parser assumes that there is HTML mixed in with the JSON
func (x *Scraper) jsonSchema4(body []byte, recipe *RecipeObject) (ok bool) {
	textOut := make([]string, 0)
	// is this a schema.org JSON string
	if !strings.Contains(string(body), "schema.org") {
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
	case x.jsonSchema1(text, recipe):
		return true
	case x.jsonSchema2(text, recipe):
		return true
	case x.jsonSchema3(text, recipe):
		return true
	}
	return false
}
