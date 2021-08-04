package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

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
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML1RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
				rawTag := string(tokenizer.Raw())
				if strings.Contains(rawTag, "ingredient") {
					textIsIngredient = true
				}
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			text = strings.TrimRight(text, "\n ")
			text = strings.TrimSpace(text)
			if textIsIngredient {
				ingredients = append(ingredients, string(text))
				textIsIngredient = false
			}
		}
	}
	return false
}

/*
<div class="wprm-recipe-ingredient-group">
<ul class="wprm-recipe-ingredients">
<li class="wprm-recipe-ingredient" style="list-style-type: disc;">
<span class="wprm-recipe-ingredient-amount">
1 1/4 </span>
<span class="wprm-recipe-ingredient-unit"> cups</span>
<span class="wprm-recipe-ingredient-name"> ground almonds  (not almond flour)</span>
<span class="wprm-recipe-ingredient-notes wprm-recipe-ingredient-notes-faded"> (130 grams)</span>
</li>
<li class="wprm-recipe-ingredient" style="list-style-type: disc;">
<span class="wprm-recipe-ingredient-amount"> 1/4</span>
<span class="wprm-recipe-ingredient-unit"> cup</span>
<span class="wprm-recipe-ingredient-name"> + 2 1/2 tablespoons granulated sugar</span>
<span class="wprm-recipe-ingredient-notes wprm-recipe-ingredient-notes-faded"> (81 grams)</span>
</li> <li class="wprm-recipe-ingredient" style="list-style-type: disc;">
<span class="wprm-recipe-ingredient-amount"> 1</span>
<span class="wprm-recipe-ingredient-unit"> large</span>
<span class="wprm-recipe-ingredient-name"> egg white (room temperature)</span>
</li> <li class="wprm-recipe-ingredient" style="list-style-type: disc;">
<span class="wprm-recipe-ingredient-amount"> 1/2</span>
<span class="wprm-recipe-ingredient-unit"> teaspoon</span>
<span class="wprm-recipe-ingredient-name"> almond or vanilla extract</span>
</li> </ul>
</div>
*/
func (x *Scraper) html2Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	rawTag := ""
	ingredientParts := make([]string,0)
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML2RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			rawTag = string(tokenizer.Raw())
			switch string(name) {
			case "span":
				switch string(name) {
				case "span":
					switch {
					case strings.Contains(rawTag, "ingredient-amount"):
					case strings.Contains(rawTag, "ingredient-unit"):
					case strings.Contains(rawTag, "ingredient-name"):
					default:
						textIsIngredient = false
						continue
					}
					textIsIngredient = true
				}
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			text = strings.TrimRight(text, "\n ")
			text = strings.TrimSpace(text)
			if textIsIngredient {
				switch {
				case strings.Contains(rawTag, "ingredient-amount"):
					ingredientParts = make([]string,0)
					ingredientParts = append(ingredientParts, text)
				case strings.Contains(rawTag, "ingredient-unit"):
					ingredientParts = append(ingredientParts, text)
				case strings.Contains(rawTag, "ingredient-name"):
					ingredientParts = append(ingredientParts, text)
					//fmt.Println(string(text))
					ingredients = append(ingredients, strings.Join(ingredientParts," "))
					textIsIngredient = false
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
				recipe.Type = HTML3RecipeType
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
				//fmt.Println(string(text))
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
<div class="chicory-default-button-container " style="
display:inline-block !important;
position:relative !important;
margin-bottom: 18px !important;
margin-top: 9px !important;
">
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
				recipe.Type = HTML4RecipeType
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
			if textIsIngredient{
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
