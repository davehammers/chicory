package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
	"fmt"
)

// flat schema.org recipe
type SchemaOrgRecipe struct {
	Name             string      `json:"name"`
	RecipeCategory   interface{} `json:"recipeCategory"`
	RecipeCuisine    interface{} `json:"recipeCuisine"`
	RecipeIngredient []string    `json:"recipeIngredient"`
	Image            interface{} `json:"image"`
}

type Image struct {
	Type *string `json:"@type"`
	ID         *string `json:"@id"`
	//	Height    int    `json:"height"`
	//	Thumbnail string `json:"thumbnail"`
	URL *string `json:"url"`
	//	Width     int    `json:"width"`
}

type idType struct {
	ID *string `json:"@id"`
}

type cuisineType struct {
	Cuisine *string `json:"recipeCuisine"`
}

// schemaOrg_RecipeJSON parse json schema 1
// http://30pepperstreet.com/recipe/endive-salad/
func (x *Scraper) schemaOrg_RecipeJSON(sourceURL string, text []byte, recipe *RecipeObject) (ok bool) {
	r := SchemaOrgRecipe{}
	err := json.Unmarshal(text, &r)
	switch err {
	case nil:
		if len(r.RecipeIngredient) == 0 {
			break
		}
		recipe.Scraper = "JSON Single schemaOrg Recipe"
		x.appendLine(recipe, r.RecipeIngredient)

		// the remaining fields are optional
		recipe.RecipeCategory = x.extractString(sourceURL, r.RecipeCategory)
		recipe.RecipeCuisine = x.extractString(sourceURL, r.RecipeCuisine)
		recipe.Image = x.extractString(sourceURL, r.Image)

		return true
	default:
		// Unmarshal error
	}
	return false
}
func (x *Scraper) extractString(sourceURL string, in interface{}) string {
	switch t := in.(type) {
	case nil:
		return ""
	case string:
		return t
	case []string:
		for _, str := range t {
			return str
		}
	case []interface{}:
		if len(t) == 0 {
			return ""
		}

		// try marshalling into a known data struct
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Sprintf("Error %s", err.Error())
		}

		// try unmarshalling different object types
		cuisine := &cuisineType{}
		err = json.Unmarshal(b, cuisine)
		switch {
		case err != nil:
		case cuisine.Cuisine == nil:
		case len(*cuisine.Cuisine) == 0:
		default:
			return *cuisine.Cuisine
		}

		// maybe a []string
		for _, v := range t {
			return fmt.Sprint(v)
		}
		return fmt.Sprintf("Error: %s %T, %q\n", sourceURL, in, in)
	case map[string]interface{}:
		b, err := json.Marshal(in)
		if err != nil {
			fmt.Println(err)
			return fmt.Sprintf("Error %s", err.Error())
		}
		// try unmarshalling different object types
		id := &idType{}
		err = json.Unmarshal(b, id)
		switch {
		case err != nil:
		case id.ID == nil:
			// not an ID struct
		default:
			return *id.ID
		}

		// try image object
		img := &Image{}
		err = json.Unmarshal(b, img)
		switch {
		case err != nil:
		case img.Type == nil:
			// it isn't an Image struct
		case *img.Type != "ImageObject":
			// it isn't an Image type
		case img.URL != nil:
			// if URL is present, return the URL
			return *img.URL
		case img.ID != nil:
			// if ID is present, return the ID
			return *img.ID
		default:
			return ""
		}
		return fmt.Sprintf("Error: Parsing %T, %q\n", in, in)

	default:
		return fmt.Sprintf("Error: extractString something unexpected %T %q", in, in)
	}
	return ""
}
