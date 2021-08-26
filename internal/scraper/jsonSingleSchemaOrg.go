package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type SchemaOrgImageObject struct {
}

// flat schema.org recipe
type SchemaOrgRecipe struct {
	Name             string      `json:"name"`
	RecipeCategory   interface{}      `json:"recipeCategory"`
	RecipeCuisine    interface{}      `json:"recipeCuisine"`
	RecipeIngredient []string    `json:"recipeIngredient"`
	Image            interface{} `json:"image"`
}

// schemaOrg_RecipeJSON parse json schema 1
// http://30pepperstreet.com/recipe/endive-salad/
func (x *Scraper) schemaOrg_RecipeJSON(siteUrl string, text []byte, recipe *RecipeObject) (ok bool) {
	r := SchemaOrgRecipe{}
	err := json.Unmarshal(text, &r)
	switch err {
	case nil:
		if len(r.RecipeIngredient) == 0 {
			break
		}
		recipe.Scraper = append(recipe.Scraper, "JSON Single schemaOrg Recipe")
		x.appendLine(recipe, r.RecipeIngredient)
		ok = true

		// the remaining fields are optional
		x.extractString(r.RecipeCategory)
		x.extractString(r.RecipeCuisine)
		x.extractString(r.Image)
	default:
		// Unmarshal error
	}
	return
}
func (x *Scraper) extractString(in interface{}) string {
	b, _ := json.MarshalIndent(in, "", "    ")
	switch t := in.(type) {
	case string:
		return t
	case []string:
		for _, str := range t {
			return str
		}
	case []interface{}:
		 for idx, inf := range t {
		 	switch infType := inf.(type) {
			case string:
				return infType
			default:
				log.Printf("%d, %T, %q\n", idx, inf, inf)
			}
		}
	default:
		log.Println("something unexpected", t)
	}
	return ""
}
