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
func (x *jsonParserType) graph_schemaOrgJSON(text string) (ok bool) {
	switch {
	case x.singleGraph_schemaOrgJSON(text):
	case x.listGraph_schemaOrgJSON(text):
	default:
		return false
	}
	return true
}
func (x *jsonParserType) singleGraph_schemaOrgJSON(text string) (ok bool) {
	r := RecipeSchema2{}
	err := json.Unmarshal([]byte(text), &r)
	if err == nil {
		for _, entry := range r.Graph {
			if b, err := json.Marshal(entry); err == nil {
				// parse individual entries for schema.org/Recipe
				if !x.schemaOrg_RecipeJSON(string(b)) {
					continue
				}
				// recipe contents was populated in the fumction above
				x.recipe.Scraper = "JSON Graph schemaOrg Recipe"
				return true
			}
		}
	}
	return false
}
func (x *jsonParserType) listGraph_schemaOrgJSON(text string) (ok bool) {
	r := RecipeSchema2List{}
	err := json.Unmarshal([]byte(text), &r)
	if err != nil {
		return false
	}
	for _, entry := range r {
		if b, err := json.Marshal(entry); err == nil {
			if ok = x.singleGraph_schemaOrgJSON(string(b)); ok {
				x.recipe.Scraper = "JSON List Graph schemaOrg Recipe"
				return
			}
		}
	}
	return false
}
