package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// schemaOrgMicrodata - schema.org Microdata format
func (x *Scraper) schemaOrgMicrodata(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	recipe.RecipeIngredient	= nil

	recipeBlock := false
	textIsIngredient := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	recipeFormat := ""
	textParts := make([]string,0)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = recipeFormat
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			raw := string(tokenizer.Raw())
			switch string(name) {
			case "div":
				if strings.Contains(raw, "https://schema.org/Recipe") {
					recipeFormat = "schemaOrgMicrodata"
					recipeBlock = true
				} else if strings.Contains(raw, "https://schema.org/")  && strings.Contains(raw, "Recipe") {
					recipeFormat = "schemaOrgRDFa"
					recipeBlock = true
				}
			case "li":
				if recipeBlock {
					if strings.Contains(raw, "recipeIngredient") {
						textIsIngredient = true
						textParts = nil
					}
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
				if textIsIngredient {
					x.appendLine(recipe, strings.Join(textParts, " "))
				}
				textIsIngredient = false
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			if textIsIngredient {
				textParts = append(textParts, text)
			}
		}
	}
	return false
}
