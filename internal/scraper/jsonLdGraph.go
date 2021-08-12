package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
)

// nexted in Graph data
type RecipeSchema2 struct {
	Context string `json:"@context"`
	Graph   []struct {
		RecipeIngredient []string `json:"recipeIngredient,omitempty"`
	} `json:"@graph"`
}

// graph_schemaOrgJSON parse json schema 2
// http://ahealthylifeforme.com/25-minute-garlic-mashed-potatoes
func (x *Scraper) graph_schemaOrgJSON(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.Graph {
			if len(entry.RecipeIngredient) > 0 {
				recipe.Scraper = append(recipe.Scraper, "JSON Graph schemaOrg Recipe")
				x.appendLine(recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}
