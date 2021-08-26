package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
)

// nexted in Graph data
type RecipeSchema2 struct {
	Context string        `json:"@context"`
	Graph   []interface{} `json:"@graph"`
}
type RecipeSchema2List []RecipeSchema2

// graph_schemaOrgJSON parse json schema 2
// http://ahealthylifeforme.com/25-minute-garlic-mashed-potatoes
func (x *Scraper) graph_schemaOrgJSON(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	switch {
	case x.singleGraph_schemaOrgJSON(siteURL, text, recipe):
	case x.listGraph_schemaOrgJSON(siteURL, text, recipe):
	default:
		return false
	}
	return true
}
func (x *Scraper) singleGraph_schemaOrgJSON(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal(text, &r)
	if err == nil {
		for _, entry := range r.Graph {
			if b, err := json.Marshal(entry); err == nil {
				// parse individual entries for schema.org/Recipe
				if !x.schemaOrg_RecipeJSON(siteURL, b, recipe) {
					continue
				}
				// recipe contents was populated in the fumction above
				recipe.Scraper = append(recipe.Scraper, "JSON Graph schemaOrg Recipe")
				return true
			}
		}
	}
	return false
}
func (x *Scraper) listGraph_schemaOrgJSON(siteURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := RecipeSchema2List{}
	err := json.Unmarshal(text, &r)
	if err != nil {
		return false
	}
	for _, entry := range r {
		if b, err := json.Marshal(entry); err == nil {
			if ok = x.singleGraph_schemaOrgJSON(siteURL, b, recipe); ok {
				recipe.Scraper[0] = "JSON List Graph schemaOrg Recipe"
				return
			}
		}
	}
	return false
}
