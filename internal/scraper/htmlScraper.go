package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

type tokenActions struct {
	addSpace bool
	end      bool
}

type HTMLScraper func(siteUrl string, body []byte, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	switch {
	case x.schemaOrgMicrodata(siteUrl, body, recipe):
	case x.htmlHRecipe(siteUrl, body, recipe):
	case x.htmlCustomDivDiv(siteUrl, body, recipe):
	case x.htmlCustomLiLi(siteUrl, body, recipe):
	case x.htmlCustomScanScan(siteUrl, body, recipe):
	case x.htmlCustomLabelLabel(siteUrl, body, recipe):
	case x.htmlCustomDivLiLiDiv(siteUrl, body, recipe):
	default:
		return false
	}
	return true
}

