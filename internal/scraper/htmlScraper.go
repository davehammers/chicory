package scraper

import (
	"golang.org/x/net/html"
)

// contains definitions and functions for accessing and parsing recipes from URLs

type tokenActions struct {
	keyWord string
	addSpace bool
	end      bool
}

//type HTMLScraper func(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool) {
	//fmt.Println(string(body))
	switch {
	case x.htmlSingleElemScraper(sourceURL, doc, recipe):
	case x.htmlNestedElemScraper(sourceURL, doc, recipe):
	default:
		return false
	}
	return true
}

//traverseNodes - interates over an HTML parsed doc invoking startNode() and endNode() callbacks for each node
func (x *Scraper) traverseNodes(n *html.Node, startNode func(*html.Node), endNode func(*html.Node)) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		startNode(c)
		x.traverseNodes(c, startNode, endNode)
		endNode(c)
	}
}