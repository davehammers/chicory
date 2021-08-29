package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

/*
  <li class="wprm-recipe-ingredient" style="list-style-type: disc;">
          <span class="wprm-recipe-ingredient-amount">2</span>
          <span class="wprm-recipe-ingredient-unit">tbsp</span>
          <span class="wprm-recipe-ingredient-name">butter</span>
          <span class="wprm-recipe-ingredient-notes wprm-recipe-ingredient-notes-faded">softened</span>
  </li>
*/
func (x *Scraper) htmlCustomScanScan(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	type tokenActions struct {
		keyWord string
		addSpace bool
		end      bool
	}
	// current action for text token
	var textAction tokenActions

	recipe.RecipeIngredient = nil
	rawTag := ""
	ingredientParts := ""
	textIsIngredient := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(recipe.RecipeIngredient) > 0 {
				recipe.Scraper = "custom <span></span>"
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "span":
				rawTag = string(tokenizer.Raw())
				for _, v := range []tokenActions{
					{"recipeingredient",                 false, true},
					{"recipeIngredient",                 false, true},
					{`class="amount"`,                   true, false},
					{`class="name"`,                     false, true},

					{"wprm-recipe-ingredient-amount",    true, false},
					{"wprm-recipe-ingredient-unit",      true, false},
					{"wprm-recipe-ingredient-name",      false, true},

					{"wpurp-recipe-ingredient-quantity", true, false},
					{"wpurp-recipe-ingredient-name",     false, true},
					{`itemprop="ingredients"`,           false, true},
				} {
					if strings.Contains(rawTag, v.keyWord) {
						textAction = v
						textIsIngredient = true
						break
					}
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			rawTag = string(tokenizer.Raw())
			switch string(name) {
			case "span":
				textIsIngredient = false
			}
		case html.TextToken:
			if textIsIngredient {
				text := string(tokenizer.Text())
				ingredientParts += text
				if textAction.addSpace {
					ingredientParts += " "
				}
				if textAction.end {
					x.appendLine(recipe, ingredientParts)
					ingredientParts = ""
				}
			}
		}
	}
	return false
}
