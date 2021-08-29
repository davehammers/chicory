package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// htmlHRecipe - micro format
func (x *Scraper) htmlHRecipe(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	recipe.RecipeIngredient = nil

	recipeBlock := false
	textIsIngredient := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	textParts := make([]string, 0)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = "hrecipe"
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			raw := string(tokenizer.Raw())
			switch string(name) {
			case "article":
				for _, token := range []string{"h-recipe", "hrecipe"} {
					if strings.Contains(raw, token) {
						recipeBlock = true
						break
					}
				}
			case "li":
				if recipeBlock {
					for _, token := range []string{"p-ingredient", `class="ingredient-item`} {
						if strings.Contains(raw, token) {
							textParts = nil
							textIsIngredient = true
						}
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
