package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

/*
<li class="ingredient" itemprop="ingredients">head of raw cauliflower, chopped into florets</li>
<li class="ingredient" itemprop="ingredients">tablespoon olive oil</li>
<li class="ingredient" itemprop="ingredients">sweet onion, chopped</li>
<li class="ingredient" itemprop="ingredients">stalks of curly kale, stem removed and chopped</li>
<li class="ingredient" itemprop="ingredients">1 teaspoon salt</li>
<li class="ingredient" itemprop="ingredients">1/2 teaspoon pepper</li>
*/
func (x *Scraper) htmlCustomLiLi(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	// current action for text token
	var textAction tokenActions

	textIsIngredient := false
	ingredients := make([]string, 0)
	ingredientParts := ""
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Scraper = "custom <li></li>"
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			rawTag := strings.ToLower(string(tokenizer.Raw()))
			switch string(name) {
			case "li":
				for _, v := range []tokenActions{
						{`class="ingredient"`,      false, true},
						{`class="ingredient `,      false, true},
						{`class="ingredients"`,     false, true},
						{`itemprop="ingredients"`,  false, true},
						{`itemprop=ingredients`,    false, true},
						{`class="ingredient-item"`, false, true},
						{"p-ingredient",            false, true},
					} {
					if strings.Contains(rawTag, v.keyWord) {
						textAction = v
						textIsIngredient = true
						break
					}
				}
			case "span":
				textIsIngredient = false
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
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
