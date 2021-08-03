package scraper
// contains definitions and functions for accessing and parsing recipes from URLs

import (
	log "github.com/sirupsen/logrus"
)

type RecipeParseType string

const (
	JSON1RecipeType RecipeParseType = "JSON1"
	JSON2RecipeType RecipeParseType = "JSON2"
	JSON3RecipeType RecipeParseType = "JSON3"
	HTML1RecipeType RecipeParseType = "HTML1"
	HTML2RecipeType RecipeParseType = "HTML2"
	HTML3RecipeType RecipeParseType = "HTML3"
	HTML4RecipeType RecipeParseType = "HTML4"
)

type RecipeObject struct {
	Type             RecipeParseType `json:"type"`
	RecipeIngredient []string        `json:"recipeIngredient"`
}

// ScrapeRecipe scrapes recipe from body returns RecipeObject,ok= true if found
func (x *Scraper) ScrapeRecipe(siteUrl string, body []byte) (recipe *RecipeObject, found bool) {
	recipe = &RecipeObject{}
	switch {
	case x.jsonParser(siteUrl, body, recipe):
		found = true
	case x.htmlParser(siteUrl, body, recipe):
		found = true
	default:
		log.Warn("No Parser match ", siteUrl)
		found = false
		return
	}
	if err := x.addRecipeToCache(siteUrl,*recipe);err != nil {
		log.Warn(err)
	}
	return
}
