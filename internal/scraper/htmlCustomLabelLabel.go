package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func (x *Scraper) htmlCustomLabelLabel(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	recipe.RecipeIngredient = nil
	tokenWords := []string{
		"recipeingredient",
	}
	textIsIngredient := false
	ingredientParts := ""
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "custom <label></label>")
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			rawTag := strings.ToLower(string(tokenizer.Raw()))
			switch string(name) {
			case "label":
				for _, v := range tokenWords {
					if strings.Contains(rawTag, v) {
						textIsIngredient = true
						break
					}
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "label":
				x.appendLine(recipe, ingredientParts)
				textIsIngredient = false
				ingredientParts = ""
			}
		case html.TextToken:
			if textIsIngredient {
				text := string(tokenizer.Text())
				ingredientParts += text
			}
		}
	}
	return false
}
