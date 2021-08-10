package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"net/http"
)

const (
	SchemaOrgRecipeType        = "SchemaOrgRecipeType"
	graph_schemaOrgJSONType    = "GraphEmbeddedSchemaOrgJSONType"
	schemaOrg_ListType         = "schemaOrg_ListType"
	schemaOrg_ItemListJSONType = "schemaOrgItemListJSONType"

	HTML1RecipeType = "<li></li>"
	HTML2RecipeType = "<scan></scan>"
	HTML3RecipeType = "HTML3"
	HTML4RecipeType = "HTML4"
	HTML5RecipeType = "HTML5"
)

type RecipeObject struct {
	SiteURL          string   `json:"url"`
	StatusCode       int      `json:"statusCode"`
	Error            string   `json:"error"`
	Scraper           []string `json:"scraper"`
	Attributes           []string `json:"attributes"`
	RecipeIngredient []string `json:"recipeIngredient"`
}

// ScrapeRecipe scrapes recipe from body returns RecipeObject, found = true if found
func (x *Scraper) ScrapeRecipe(siteURL string, body []byte) (recipe *RecipeObject, found bool) {
	//os.WriteFile("dump.html", body, 0666)
	recipe = &RecipeObject{
		SiteURL:    siteURL,
		StatusCode: http.StatusOK,
		Error:      "",
	}
	jsonFound := x.jsonParser(siteURL, body, recipe)
	httpFound := x.htmlParser(siteURL, body, recipe)
	if jsonFound || httpFound {
		found = true
		return
	}
	recipe.SiteURL = siteURL
	recipe.StatusCode = http.StatusNotImplemented
	recipe.Error = "No Parser match"
	found = false
	return
}
