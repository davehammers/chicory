package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
)

// top level list
type RecipeSchema3 []struct {
	Context          string   `json:"@context"`
	Type             string   `json:"@type"`
	RecipeIngredient []string `json:"recipeIngredient,omitempty"`
}

// schemaOrg_List parse json schema 3
// http://allrecipes.com/recipe/12646/cheese-and-garden-vegetable-pie/
func (x *Scraper) schemaOrg_List(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema3{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "JSON List schemaOrg Recipe")
				x.appendLine(recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}
