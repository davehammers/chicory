package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
)

type RecipeSchema4 struct {
	Context         string `json:"@context"`
	Type            string `json:"@type"`
	ItemListElement []struct {
		RecipeIngredient []string `json:"recipeIngredient"`
	} `json:"itemListElement"`
	NumberOfItems int    `json:"numberOfItems"`
	ItemListOrder string `json:"itemListOrder"`
}

// schemaOrg_ItemListJSON parse json schema 4
// https://www.yummly.com/recipe/Roasted-garlic-caesar-dipping-sauce-297499
func (x *jsonParserType) schemaOrg_ItemListJSON(text string) (ok bool) {
	r := RecipeSchema4{}
	err := json.Unmarshal([]byte(text), &r)
	if err == nil {
		for _, entry := range r.ItemListElement {
			if len(entry.RecipeIngredient) > 0 {
				x.recipe.Scraper = "JSON schemaOrg ItemList Recipe"
				x.scraper.appendLine(x.recipe,  entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}
