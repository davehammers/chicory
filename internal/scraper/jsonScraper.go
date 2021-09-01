package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"golang.org/x/net/html/atom"
	"strings"

	"golang.org/x/net/html"
)

const (
	RecipeType = "Recipe"
	LdType     = "@type"
)

type jsonParserType struct {
	insideScript bool
	sourceURL string
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
			case x.schemaOrg_ItemListJSON(n.Data):
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
