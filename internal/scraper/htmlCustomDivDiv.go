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
	divBlock := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	textParts := make([]string, 0)
	spanCnt := 0
	divWords := []string{
		`"ingredient"`,
	}
	spanWords := []string{
		`"recipeIngredient"`,
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
						divBlock = true
						break
					}
				}
			case "span":
				if !divBlock{
					break
				}
				for _, word := range spanWords {
					if strings.Contains(raw, word) {
						textIsIngredient = true
						spanCnt = 0
						break
					}
				}
				spanCnt++
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "span":
				if textIsIngredient {
					spanCnt--
					if spanCnt == 0 {
						x.appendLine(recipe, strings.Join(textParts, ""))
						textParts = nil
						textIsIngredient = false
						divBlock = false
					}
				}
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
