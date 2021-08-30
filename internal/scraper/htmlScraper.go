package scraper

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// contains definitions and functions for accessing and parsing recipes from URLs

type tokenActions struct {
	keyWord string
	addSpace bool
	end      bool
}

//type HTMLScraper func(sourceURL string, doc *html.Node, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	//fmt.Println(string(body))
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		log.Error(err)
		return false
	}
	switch {
	case x.htmlSingleElemScraper(sourceURL, doc, recipe):
	case x.htmlNestedElemScraper(sourceURL, doc, recipe):
	//case x.schemaOrgMicrodata(sourceURL, body, recipe):
	//case x.htmlHRecipe(sourceURL, body, recipe):
	//case x.htmlCustomDivDiv(sourceURL, body, recipe):
	//case x.htmlCustomLiLi(sourceURL, body, recipe):
	//case x.htmlCustomScanScan(sourceURL, body, recipe):
	//case x.htmlCustomLabelLabel(sourceURL, body, recipe):
	//case x.htmlCustomDivLiLiDiv(sourceURL, body, recipe):
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