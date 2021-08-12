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
func (x *Scraper) htmlCustomScanScan(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	type tokenActions struct {
		addSpace bool
		end      bool
	}
	tokenWords := map[string]tokenActions{
		"recipeIngredient":                 {false, true},
		`class="amount"`:                   {true, false},
		`class="name"`:                     {false, true},
		"wprm-recipe-ingredient-amount":    {true, false},
		"wprm-recipe-ingredient-unit":      {true, false},
		"wprm-recipe-ingredient-name":      {false, true},
		"wpurp-recipe-ingredient-quantity": {true, false},
		"wpurp-recipe-ingredient-name":     {false, true},
	}
	// current action for text token
	var textAction tokenActions

	rawTag := ""
	ingredientParts := ""
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Scraper = append(recipe.Scraper, "custom <scan></scan>")
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			rawTag = string(tokenizer.Raw())
			switch string(name) {
			case "span":
				for k, v := range tokenWords {
					if strings.Contains(rawTag, k) {
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
				text = strings.TrimRight(text, "\n ")
				text = strings.TrimSpace(text)
				if text == "" {
					break
				}
				ingredientParts += text
				if textAction.addSpace {
					ingredientParts += " "
				}
				if textAction.end {
					ingredients = append(ingredients, ingredientParts)
					ingredientParts = ""
				}
			}
		}
	}
	return false
}
