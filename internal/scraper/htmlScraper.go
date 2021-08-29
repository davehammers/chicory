package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

type tokenActions struct {
	keyWord string
	addSpace bool
	end      bool
}

type HTMLScraper func(sourceURL string, body []byte, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(sourceURL string, body []byte, recipe *RecipeObject) (found bool) {
	switch {
	case x.htmlTokenizer(sourceURL, body, recipe):
	case x.schemaOrgMicrodata(sourceURL, body, recipe):
	case x.htmlHRecipe(sourceURL, body, recipe):
	case x.htmlCustomDivDiv(sourceURL, body, recipe):
	case x.htmlCustomLiLi(sourceURL, body, recipe):
	case x.htmlCustomScanScan(sourceURL, body, recipe):
	case x.htmlCustomLabelLabel(sourceURL, body, recipe):
	case x.htmlCustomDivLiLiDiv(sourceURL, body, recipe):
	default:
		return false
	}
	return true
}

