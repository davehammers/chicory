package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

/*
<div class="mv-create-ingredients">
<li>
2 1/2 pounds of cine-ripened tomatoes (or canned, chopped tomatoes)					</li>
<li>
</div>
*/
func (x *Scraper) htmlCustomDivLiLiDiv(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	recipe.RecipeIngredient = nil
	recipeBlock := false
	textIsIngredient := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	textParts := make([]string, 0)
	divWords := []string{
		"ingredients",
		"data-tasty-recipes-customization",
	}

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = "custom <div><li></li></div>"
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			raw := string(tokenizer.Raw())
			switch string(name) {
			case "div":
				for _, word := range divWords {
					if strings.Contains(raw, word) {
						textIsIngredient = true
						recipeBlock = true
						break
					}

				}
				if strings.Contains(strings.ToLower(raw), "ingredients") {
					textIsIngredient = true
					recipeBlock = true
				}
			case "li":
				if recipeBlock {
					textIsIngredient = true
					textParts = nil
				}
			}

		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				recipeBlock = false
				textIsIngredient = false
			case "li":
				if textIsIngredient {
					x.appendLine(recipe, strings.Join(textParts, ""))
					textParts = nil
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
