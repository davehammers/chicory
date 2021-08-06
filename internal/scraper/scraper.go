package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"net/http"
)

type RecipeParseType string

const (
	SchemaOrgRecipeType        RecipeParseType = "SchemaOrgRecipeType"
	graph_schemaOrgJSONType    RecipeParseType = "embeddedGraphSchemaOrgJSONType"
	schemaOrg_ListType         RecipeParseType = "schemaOrg_ListType"
	schemaOrg_ItemListJSONType RecipeParseType = "schemaOrgItemListJSONType"

	HTML1RecipeType RecipeParseType = "HTML1"
	HTML2RecipeType RecipeParseType = "HTML2"
	HTML3RecipeType RecipeParseType = "HTML3"
	HTML4RecipeType RecipeParseType = "HTML4"
	HTML5RecipeType RecipeParseType = "HTML5"
)

type RecipeObject struct {
	SiteURL          string
	StatusCode       int
	Error            string
	Type             RecipeParseType `json:"type"`
	RecipeIngredient []string        `json:"recipeIngredient"`
}

// ScrapeRecipe scrapes recipe from body returns RecipeObject, found = true if found
func (x *Scraper) ScrapeRecipe(siteUrl string, body []byte) (recipe *RecipeObject, found bool) {
	//os.WriteFile("dump.html", body, 0666)
	recipe = &RecipeObject{
		SiteURL: siteUrl,
		StatusCode: http.StatusOK,
		Error: "",
	}
	switch {
	case x.jsonParser(siteUrl, body, recipe):
		found = true
	case x.htmlParser(siteUrl, body, recipe):
		found = true
	default:
		recipe.StatusCode = http.StatusNotImplemented
		recipe.Error = "No Parser match"
		found = false
		return
	}
	return
}
