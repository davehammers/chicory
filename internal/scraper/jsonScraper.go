package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/html/atom"

	"golang.org/x/net/html"
)

// top level list
type RecipeSchema3 []struct {
	Context          string   `json:"@context"`
	Type             string   `json:"@type"`
	RecipeIngredient []string `json:"recipeIngredient,omitempty"`
}

// nexted in Graph data
type RecipeSchema2 struct {
	Context string        `json:"@context"`
	Graph   []interface{} `json:"@graph"`
}

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
	ID   *string `json:"@id"`
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

// jsonParserType - control structure for JSON parsers
type jsonParserType struct {
	insideScript bool
	sourceURL    string
	recipe       *RecipeObject
	scraper      *Scraper
	rebuildText  []string
}

func (x *Scraper) NewJsonScraper(recipe *RecipeObject) *jsonParserType {
	return &jsonParserType{
		insideScript: false,
		recipe:       recipe,
		scraper:      x,
	}
}

// jsonParser tries to extract recipe in JSON-LD format
func (x *Scraper) jsonParser(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool) {
	// traverse the HTML nodes calling startNode at the beginning and endNode at the end of each node
	j := x.NewJsonScraper(recipe)
	j.sourceURL = sourceURL
	x.traverseNodes(doc, j.jsonStartNode, j.jsonEndNode)
	if len(recipe.RecipeIngredient) > 0 {
		return true
	}
	return false
}

func (x *jsonParserType) jsonStartNode(n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		switch n.DataAtom {
		case atom.Script:
			x.insideScript = true
		default:
			x.insideScript = false
		}
	case html.TextNode:
		if !x.insideScript {
			return
		}
		if !strings.Contains(string(n.Data), "schema.org") {
			return
		}
		for retry := 0; retry < 2; retry++ {
			switch {
			case x.graph_schemaOrgJSON(n.Data):
			case x.schemaOrg_RecipeJSON(n.Data):
			case x.schemaOrg_List(n.Data):
			case x.jsonSchemaRemoveHTML(n.Data):
				// try to strip the text of any embedded HTML and try again
				n.Data = strings.Join(x.rebuildText, " ")
				continue
			}
			// match found
			if retry > 0 {
				x.recipe.Error = "JSON invalid. Attempting to repair"
			}
			break
		}
	}
}

func (x *jsonParserType) jsonEndNode(n *html.Node) {
	x.insideScript = false
}

// jsonSchemaRemoveHTML parse json schema RemoveHTML
// this parser assumes that there is HTML mixed in with the JSON
// It tries to remove all of the mixed in HTML then reprocess the JSON
// https://mealpreponfleek.com/low-carb-hamburger-helper/
func (x *jsonParserType) jsonSchemaRemoveHTML(body string) (ok bool) {
	x.rebuildText = make([]string, 0)
	// is this a schema.org JSON string
	if !strings.Contains(string(body), "recipeIngredient") {
		return false
	}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		// cannot parse script as HTML string
		return false
	}
	x.scraper.traverseNodes(doc, x.jsonStartFixer, x.jsonEndFixer)
	// return true to cause the other parse routines to try again
	return true
}

//jsonStartFixer - called at the beginning of each HTML element
func (x *jsonParserType) jsonStartFixer(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		// consolidate all of the text nodes into a single string
		text := n.Data
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")
		text = strings.TrimSpace(text)
		if len(text) > 0 {
			x.rebuildText = append(x.rebuildText, text)
		}
	}
}

//jsonEndFixer - called at the end of each HTML node
func (x *jsonParserType) jsonEndFixer(n *html.Node) {
}

// schemaOrg_RecipeJSON parse json schema 1
// http://30pepperstreet.com/recipe/endive-salad/
func (x *jsonParserType) schemaOrg_RecipeJSON(text string) (ok bool) {
	r := SchemaOrgRecipe{}
	err := json.Unmarshal([]byte(text), &r)
	switch err {
	case nil:
		if len(r.RecipeIngredient) == 0 {
			break
		}
		x.recipe.Scraper = "JSON Single schemaOrg Recipe"
		x.scraper.appendLine(x.recipe, r.RecipeIngredient)

		// the remaining fields are optional
		x.recipe.Name = x.extractString(r.Name)
		x.recipe.RecipeCategory = x.extractString(r.RecipeCategory)
		x.recipe.RecipeCuisine = x.extractString(r.RecipeCuisine)
		x.recipe.Image = x.extractString(r.Image)

		return true
	default:
		// Unmarshal error
	}
	return false
}

// graph_schemaOrgJSON parse json schema 2
// http://ahealthylifeforme.com/25-minute-garlic-mashed-potatoes
func (x *jsonParserType) graph_schemaOrgJSON(text string) (ok bool) {
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

// schemaOrg_List parse json schema 3
// http://allrecipes.com/recipe/12646/cheese-and-garden-vegetable-pie/
func (x *jsonParserType) schemaOrg_List(text string) (ok bool) {
	r := RecipeSchema3{}
	err := json.Unmarshal([]byte(text), &r)
	if err == nil {
		for _, entry := range r {
			if len(entry.RecipeIngredient) > 0 {
				x.recipe.Scraper = "JSON List schemaOrg Recipe"
				x.scraper.appendLine(x.recipe, entry.RecipeIngredient)
				return true
			}
		}
	}
	return false
}

// extractString - determines interface object type and extracts a string
func (x *jsonParserType) extractString(in interface{}) string {
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
		return fmt.Sprintf("Error: %s %T, %q\n", x.sourceURL, in, in)
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
