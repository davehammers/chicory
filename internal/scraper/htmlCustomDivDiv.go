package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

/*
<span id="zlrecipe-ingredients-list">
<div id="zlrecipe-ingredient-0" class="ingredient" itemprop="ingredients">3 c. candy corn
</div><div id="zlrecipe-ingredient-1" class="ingredient" itemprop="ingredients">1½ c. peanut butter
</div><div id="zlrecipe-ingredient-2" class="ingredient" itemprop="ingredients">2 c. (12 oz) chocolate chips </div></span>
*/
func (x *Scraper) htmlCustomDivDiv(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	recipe.RecipeIngredient = nil
	textIsIngredient := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	textParts := make([]string, 0)
	divWords := []string{
		"ingredient-text",
	}
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "custom <div></div>")
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
						break
					}
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				if textIsIngredient {
					x.appendLine(recipe, strings.Join(textParts, ""))
					textParts = nil
				}
				textIsIngredient = false
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			if textIsIngredient {
				x.appendLine(recipe, text)
			}
		}
	}
	return false
}
