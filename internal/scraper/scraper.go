package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"net/http"
	"strings"
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
	htmlRecipe := &RecipeObject{
		SiteURL:    siteURL,
		StatusCode: http.StatusOK,
		Error:      "",
	}
	httpFound := x.htmlParser(siteURL, body, htmlRecipe)
	if jsonFound || httpFound {
		found = true
		recipe.Scraper = append(recipe.Scraper, htmlRecipe.Scraper...)
		if len(recipe.RecipeIngredient) == 0 {
			recipe.RecipeIngredient = htmlRecipe.RecipeIngredient
		}

		return
	}
	recipe.SiteURL = siteURL
	recipe.StatusCode = http.StatusNotImplemented
	recipe.Error = "No Parser match"
	found = false
	return
}

// appendLine -  clean up a recipe line before adding it to the list
func (x *Scraper) appendLine(recipe *RecipeObject, item interface{}) {
	switch i := item.(type) {
	case string:
		x.appendString(recipe, i)
	case []string:
		for _, text := range i {
			x.appendString(recipe, text)
		}
	}
	return
}
func (x *Scraper) appendString(recipe *RecipeObject, text string) {
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "<p>", "")
	text = strings.ReplaceAll(text, "</p>", "")
	text = strings.TrimSpace(text)
	if text != "" {
		recipe.RecipeIngredient = append(recipe.RecipeIngredient, text)
	}
}
