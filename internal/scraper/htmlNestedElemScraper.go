package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type nestedElem struct {
	DataAtom    atom.Atom // node type such as <span>
	AttrKey     string    // matching attribute key
	AttrValue   string    // matching attribute value
	subDataAtom atom.Atom
	breakOnBr   bool
	isEnd       bool
}

// this table controls the behavior of a nested element ingredient
// A nested element is when the HTML element has key words that tell us the text portions are the ingredient
var nestedElements []nestedElem = []nestedElem{
	{atom.Div, "class", "mv-create-ingredients", atom.Li, false, true},
	{atom.Div, "class", "recipe__list recipe__list--ingredients", atom.Li, false, true},
	{atom.Div, "class", "wprm-fallback-recipe-ingredients", atom.Li, false, true},
	{atom.Div, "class", "tasty-recipes-ingredients", atom.Li, false, true},
	{atom.Div, "class", "tasty-recipes-ingredients", atom.P, false, true},
	{atom.Div, "class", "tasty-recipe-ingredients", atom.Li, false, true},
	{atom.Div, "class", "ccm-section-ingredients ingredients", atom.Li, false, true},
	{atom.Div, "class", "ingredients", atom.Li, false, true},
	{atom.Div, "class", "ERSIngredients", atom.Li, false, true},
	{atom.Div, "id", "recbody", atom.Li, false, true},
	{atom.Div, "class", "container container-sm", atom.Li, false, true},
	{atom.Div, "class", "penci-recipe-ingredients penci-recipe-ingre-visual", atom.P, false, true},
	{atom.Div, "class", "recipe__ingredients", atom.Li, true, true},
	{atom.Div, "class", "ingredients ingredient", atom.P, true, true},
	{atom.Div, "class", "ingredient-list__steps", atom.Li, false, true},
}

type nestedElemScraper struct {
	scraper          *Scraper
	isIngredientText bool
	text             string // working text string
	elems            []nestedElem
	curElem          *nestedElem
	curSubNode       *html.Node
	recipe           *RecipeObject
}

func (x *Scraper) htmlNestedElemScraper(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool) {
	s := x.NewNestedElemScraper(recipe)
	// traverse the HTML nodes calling startNode at the beginning and endNode at the end of each node
	x.traverseNodes(doc, s.startNode, s.endNode)
	if len(recipe.RecipeIngredient) > 0 {
		recipe.Scraper = "HTML Nested Elem Scraper"
		return true
	}
	return false
}

func (x *Scraper) NewNestedElemScraper(recipe *RecipeObject) *nestedElemScraper {
	return &nestedElemScraper{
		scraper:          x,
		isIngredientText: false,
		text:             "",
		elems:            nestedElements,
		recipe:           recipe,
	}
}

// startNode - is called for each HTML node
func (x *nestedElemScraper) startNode(n *html.Node) {
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
					// if strings.Contains(strings.ToLower(attr.Val), elem.AttrValue) {
					if strings.ToLower(attr.Val) == elem.AttrValue {
						//fmt.Printf("outer node %#v\n",n)
						//fmt.Printf("outer elem %#v\n",elem)
						x.curElem = &elem
						// process the HTML subnodes with the rules to see if they work
						x.scraper.traverseNodes(n, x.startSubNode, x.endSubNode)
						if len(x.recipe.RecipeIngredient) > 0 {
							return
						}
					}
				}
			}
		}
	}
}

func (x *nestedElemScraper) endNode(n *html.Node) {}

func (x *nestedElemScraper) startSubNode(n *html.Node) {
	//fmt.Printf("inner node %#v\n",n)
	switch n.Type {
	case html.ElementNode:
		//fmt.Printf("inner node %#v\n",n)
		//fmt.Printf("inner elem %#v\n",x.curElem)

		// handle the case where a recipe is within a single element and the lines are
		// separated by <br>
		if x.curElem.breakOnBr && n.DataAtom == atom.Br && x.isIngredientText {
			x.recipe.AppendLine(x.text)
			x.text = ""
		}
		// does the element have matching attributes
		if !x.nodeMatchesSubElement(n, x.curElem) {
			return
		}
		// next text node is in the ingredient
		x.isIngredientText = true
		x.curSubNode = n
	case html.TextNode:
		if !x.isIngredientText {
			break
		}
		x.text += n.Data
		if x.curElem.isEnd {
			x.endSubNode(n)
		}
	}
}
func (x *nestedElemScraper) nodeMatchesSubElement(n *html.Node, elem *nestedElem) bool {
	if elem == nil {
		return false
	}
	if n.DataAtom == elem.subDataAtom {
		return true
	}
	return false
}
func (x *nestedElemScraper) endSubNode(n *html.Node) {
	switch {
	case !x.isIngredientText:
	case x.curSubNode != n:
	case !x.curElem.isEnd:
	default:
		x.isIngredientText = false
		// add it to the recipe
		x.recipe.AppendLine(x.text)
		x.text = ""
		x.curSubNode = nil

	}
}
