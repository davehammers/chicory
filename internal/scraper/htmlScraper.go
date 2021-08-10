package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

type tokenActions struct {
	addSpace bool
	end      bool
}

type HTMLScraper func(siteUrl string, body []byte, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	switch {
	case x.html1Scraper(siteUrl, body, recipe):
	case x.html2Scraper(siteUrl, body, recipe):
	case x.html3Scraper(siteUrl, body, recipe):
	case x.html4Scraper(siteUrl, body, recipe):
	default:
		return false
	}
	return true
}

/*
<li class="ingredient" itemprop="ingredients">head of raw cauliflower, chopped into florets</li>
<li class="ingredient" itemprop="ingredients">tablespoon olive oil</li>
<li class="ingredient" itemprop="ingredients">sweet onion, chopped</li>
<li class="ingredient" itemprop="ingredients">stalks of curly kale, stem removed and chopped</li>
<li class="ingredient" itemprop="ingredients">1 teaspoon salt</li>
<li class="ingredient" itemprop="ingredients">1/2 teaspoon pepper</li>
*/
func (x *Scraper) html1Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	tokenWords := map[string]tokenActions{
		`class="ingredient"`:               {false, true},
		`class="ingredient `:               {false, true},
		`class="ingredients"`:              {false, true},
		`itemprop="ingredients"`:           {false, true},
		`itemprop=ingredients`:              {false, true},
		`class="ingredient-item"`:          {false, true},
	}
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
				recipe.Scraper = append(recipe.Scraper,"<li></li>")
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
				rawTag := strings.ToLower(string(tokenizer.Raw()))
				for k, v := range tokenWords {
					if strings.Contains(rawTag, k) {
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

/*
  <li class="wprm-recipe-ingredient" style="list-style-type: disc;">
          <span class="wprm-recipe-ingredient-amount">2</span>
          <span class="wprm-recipe-ingredient-unit">tbsp</span>
          <span class="wprm-recipe-ingredient-name">butter</span>
          <span class="wprm-recipe-ingredient-notes wprm-recipe-ingredient-notes-faded">softened</span>
  </li>
*/
func (x *Scraper) html2Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	type tokenActions struct {
		addSpace bool
		end      bool
	}
	tokenWords := map[string]tokenActions{
		"recipeIngredient":              {false, true},
		`class="amount"`:                {true, false},
		`class="name"`:                  {false, true},
		"wprm-recipe-ingredient-amount": {true, false},
		"wprm-recipe-ingredient-unit":   {true, false},
		"wprm-recipe-ingredient-name":   {false, true},
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
				recipe.Scraper = append(recipe.Scraper,"<scan></scan>")
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

/*
<span id="zlrecipe-ingredients-list">
<div id="zlrecipe-ingredient-0" class="ingredient" itemprop="ingredients">3 c. candy corn
</div><div id="zlrecipe-ingredient-1" class="ingredient" itemprop="ingredients">1½ c. peanut butter
</div><div id="zlrecipe-ingredient-2" class="ingredient" itemprop="ingredients">2 c. (12 oz) chocolate chips </div></span>
*/
func (x *Scraper) html3Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Scraper = append(recipe.Scraper,"<div></div>")
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "span":
				raw := string(tokenizer.Raw())
				if strings.Contains(raw, "ingredients") {
					textIsIngredient = true
				}
			}
		case html.TextToken:
			text := tokenizer.Text()
			if textIsIngredient {
				ingredients = append(ingredients, string(text))
				textIsIngredient = false
			}
		}
	}
	return false
}

/*
<div class="mv-create-ingredients">
<h3 class="mv-create-ingredients-title mv-create-title-secondary">Ingredients</h3>

<h4></h4>
<ul>
<li>
2 1/2 pounds of cine-ripened tomatoes (or canned, chopped tomatoes)					</li>
<li>
1 medium green bell pepper, chopped					</li>
<li>
1 medium red bell pepper, chopped					</li>
<li>
1 small onion, chopped					</li>
<li>
1 medium cucumber, peeled, seeded and chopped					</li>
<li>
2 large garlic cloves, minced and mashed into a paste					</li>
<li>
3 seeded, jalapeños, chopped					</li>
<li>
3 TB. red-wine vinegar					</li>
<li>
1 TB. extra virgin olive oil					</li>
<li>
Salt &amp; Pepper to taste					</li>
<li>
24 oz. tomato Juice (or ice water) for thinning					</li>
</ul>
<h4>Garnish with:</h4>
<ul>
<li>
croutons					</li>
<li>
sliced avocado					</li>
<li>
diced cucumber					</li>
</ul>
<div class="chicory-noprint" data-position="within">
<div class="chicory-default-button-container " style=" display:inline-block !important; position:relative !important;
margin-bottom: 18px !important; margin-top: 9px !important;">
*/
func (x *Scraper) html4Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Scraper = append(recipe.Scraper,"<div><li></li></div>")
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				raw := string(tokenizer.Raw())
				if strings.Contains(raw, "ingredients") {
					textIsIngredient = true
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				textIsIngredient = false
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			if textIsIngredient {
				text = strings.TrimRight(text, "\n ")
				text = strings.TrimSpace(text)
				if text == "" {
					continue
				}
				if text == "Ingredients" {
					continue
				}
				ingredients = append(ingredients, text)
			}
		}
	}
	return false
}
