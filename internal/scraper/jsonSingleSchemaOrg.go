package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
)


// flat schema.org recipe
type RecipeSchema1 struct {
	RecipeIngredient []string `json:"recipeIngredient"`
}

// schemaOrg_RecipeJSON parse json schema 1
// http://30pepperstreet.com/recipe/endive-salad/
func (x *Scraper) schemaOrg_RecipeJSON(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema1{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		if len(r.RecipeIngredient) > 0 {
			recipe.Scraper = append(recipe.Scraper, "JSON Single schemaOrg Recipe")
			x.appendLine(recipe, r.RecipeIngredient)
			return true
		}
	}
	return false
}
