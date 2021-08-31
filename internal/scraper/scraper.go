package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"fmt"
	"golang.org/x/net/html"
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
	StatusCode       int      `json:"statusCode"`
	Error            string   `json:"error"`
	SourceURL        string   `json:"url"`
	Scraper          string   `json:"scraper"`
	RecipeCategory   string   `json:"recipeCategory"`
	RecipeCuisine    string   `json:"recipeCuisine"`
	RecipeIngredient []string `json:"recipeIngredient"`
	Image            string   `json:"image"`
}

// ScrapeRecipe scrapes recipe from body returns RecipeObject, found = true if found
func (x *Scraper) ScrapeRecipe(sourceURL string, body []byte) (recipe *RecipeObject, found bool) {
	//os.WriteFile("dump.html", body, 0666)
	recipe = &RecipeObject{
		SourceURL:  sourceURL,
		StatusCode: http.StatusOK,
		Error:      "",
	}
	jsonFound := x.jsonParser(sourceURL, body, recipe)
	if jsonFound {
		found = true
		return
	}

	htmlRecipe := &RecipeObject{
		SourceURL:  sourceURL,
		StatusCode: http.StatusOK,
		Error:      "",
	}
	httpFound := x.htmlParser(sourceURL, body, htmlRecipe)
	if jsonFound || httpFound {
		found = true
		recipe.Scraper = htmlRecipe.Scraper
		if len(recipe.RecipeIngredient) == 0 {
			recipe.RecipeIngredient = htmlRecipe.RecipeIngredient
		}

		return
	}
	recipe.SourceURL = sourceURL
	if recipe.StatusCode == http.StatusOK {
		recipe.StatusCode = http.StatusUnprocessableEntity
		recipe.Error = "No Scraper Found"
		recipe.Scraper = "No Scraper Found"
	} else {
		recipe.Scraper = fmt.Sprintf("HTTP %d", recipe.StatusCode)
	}
	recipe.StatusCode = http.StatusUnprocessableEntity
	recipe.Error = "No Scraper Found"
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
	text = strings.ReplaceAll(text, "  ", " ")
	//text = x.LineRegEx.ReplaceAllLiteralString(text, " ")
	text = html.UnescapeString(text)
	text = x.AngleRegEx.ReplaceAllLiteralString(text, "")
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	if text != "" {
		recipe.RecipeIngredient = append(recipe.RecipeIngredient, text)
	}
}
// appendLine -  clean up a recipe line before adding it to the list
func (x *RecipeObject) AppendLine(item interface{}) {
	switch i := item.(type) {
	case string:
		x.AppendString(i)
	case []string:
		for _, text := range i {
			x.AppendString(text)
		}
	}
	return
}
func (x *RecipeObject) AppendString(text string) {
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "  ", " ")
	//text = x.LineRegEx.ReplaceAllLiteralString(text, " ")
	text = html.UnescapeString(text)
	//text = x.AngleRegEx.ReplaceAllLiteralString(text, "")
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	if text != "" {
		x.RecipeIngredient = append(x.RecipeIngredient, text)
	}
}
