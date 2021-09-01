package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
)

type singleElem struct {
	DataAtom  atom.Atom // node type such as <span>
	AttrKey   string    // matching attribute key
	AttrValue string    // matching attribute value
	addSpace  bool      // when aggregating the text components, add a trailing space
	isEnd     bool      // the text ends with this parameter
}

// this table controls the behavior of a single element ingredient
// A single element is when the HTML element has key words that tell us the text portions are the ingredient
var singleElements []singleElem = []singleElem{
	// <span>
	{atom.Span, "class", "wprm-recipe-ingredient-amount", true, false},
	{atom.Span, "class", "wprm-recipe-ingredient-unit", true, false},
	{atom.Span, "class", "wprm-recipe-ingredient-name", false, true},
	{atom.Span, "class", "wpurp-recipe-ingredient-quantity", true, false},
	{atom.Span, "class", "wpurp-recipe-ingredient-unit", true, false},
	{atom.Span, "class", "wpurp-recipe-ingredient-name", false, true},
	{atom.Span, "itemprop", "ingredients", false, true},
	{atom.Span, "itemprop", "recipeingredient", false, true},
	// these are too generic and match too many things
	//{atom.Span, "class", "name", false, true},
	//{atom.Span, "class", "amount", true, false},

	// <li>
	{atom.Li, "itemprop", "ingredients", false, true},
	{atom.Li, "class", "ingredient", false, true},
	{atom.Li, "itemprop", "recipeingredient", false, true},
	{atom.Li, "class", "blog-yumprint-ingredient-item", false, true},

	// <div>
	{atom.Div, "itemprop", "recipeingredient", false, true},
	{atom.Div, "class", "p-ingredient", false, true},
	{atom.Div, "itemprop", "ingredient", false, true},
	//{atom.Div, "class", "post-body entry-content", false, true},

	// <label>
	{atom.Label, "itemprop", "recipeingredient", false, true},

	// <p>
	{atom.P, "class", "ingredient", false, true},
	//{atom.P, "itemprop", "recipeingredient", false, true},
}

type singleElemScraper struct {
	isIngredientText bool
	text             string // working text string
	elems            []singleElem
	curElem          *singleElem
	curNode          *html.Node
	recipe           *RecipeObject
}

func (x *Scraper) htmlSingleElemScraper(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool) {
	// parse body into a node tree
	s := NewSingleElemScraper(recipe)
	// traverse the HTML nodes calling startNode at the beginning and endNode at the end of each node
	x.traverseNodes(doc, s.startNode, s.endNode)
	if len(recipe.RecipeIngredient) > 0 {
		recipe.Scraper = "HTML Single Elem Scraper"
		return true
	}
	return false
}

func NewSingleElemScraper(recipe *RecipeObject) *singleElemScraper {
	return &singleElemScraper{
		isIngredientText: false,
		text:             "",
		elems:            singleElements,
		recipe:           recipe,
	}
}

// startNode - is called for each HTML node
func (x *singleElemScraper) startNode(n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		// scan through the element types and look for matching criteria
		for _, elem := range x.elems {
			// does the element type match E.g. <li> or <span>
			if n.DataAtom == elem.DataAtom {
				// does the element have matching attributes
				for _, attr := range n.Attr {
					// convert element attributes to lowercase while comparing
					if attr.Key != elem.AttrKey {
						continue
					}
					if strings.Contains(strings.ToLower(attr.Val), elem.AttrValue) {
						// next text node is in the ingredient
						x.isIngredientText = true
						// curElem tells us how to handle the text
						x.curElem = &elem
						x.curNode = n
						//fmt.Printf("match Elem %#v\n", x.curElem)
						//fmt.Printf("match Node %#v\n", x.curNode)
						return
					}
				}
			}
		}
	case html.TextNode:
		if !x.isIngredientText {
			break
		}
		x.text += n.Data
		if x.curElem.addSpace {
			x.text += " "
		}
		if x.curElem.isEnd {
			x.endNode(n)
		}
	}
}

func (x *singleElemScraper) endNode(n *html.Node) {
	switch {
	case !x.isIngredientText:
	case x.curNode != n:
	case x.curElem == nil:
		x.isIngredientText = false
	case !x.curElem.isEnd:
	default:
		x.isIngredientText = false
		// add it to the recipe
		x.recipe.AppendLine(x.text)
		x.text = ""
		x.curElem = nil
		x.curNode = nil

	}
}
