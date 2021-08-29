package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

const (
	RecipeType = "Recipe"
	LdType     = "@type"
)

// jsonParser tries to extract recipe in JSON-LD format
func (x *Scraper) jsonParser(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	insideScript := false
	if strings.Contains(string(body), `"@type":"Recipe"`) ||
		strings.Contains(string(body), `"@type": "Recipe"`) {
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
			case x.graph_schemaOrgJSON(sourceURL, text, recipe):
				return true
			case x.schemaOrg_RecipeJSON(sourceURL, text, recipe):
				return true
			case x.schemaOrg_List(sourceURL, text, recipe):
				return true
			case x.schemaOrg_ItemListJSON(sourceURL, text, recipe):
				return true
			case x.jsonSchemaRemoveHTML(sourceURL, text, recipe):
				return true
			default:
				continue
			}
		}
	}
	return false
}

// jsonSchemaRemoveHTML parse json schema RemoveHTML
// this parser assumes that there is HTML mixed in with the JSON
// It tries to remove all of the mixed in HTML then reprocess the JSON
// https://mealpreponfleek.com/low-carb-hamburger-helper/
func (x *Scraper) jsonSchemaRemoveHTML(sourceURL string, body []byte, recipe *RecipeObject) (ok bool) {
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
	switch {
	case x.schemaOrg_RecipeJSON(sourceURL, text, recipe):
		return true
	case x.graph_schemaOrgJSON(sourceURL, text, recipe):
		return true
	case x.schemaOrg_List(sourceURL, text, recipe):
		return true
	case x.schemaOrg_ItemListJSON(sourceURL, text, recipe):
		return true
	}
	log.Error(sourceURL, "body contains schema.org/recipeIngredient json but did not parse")
	return false
}
