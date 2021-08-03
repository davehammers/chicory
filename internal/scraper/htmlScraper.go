package scraper

// contains definitions and functions for accessing and parsing recipes from URLs

import (
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

type HTMLScraper func(siteUrl string, body []byte, recipe *RecipeObject) (found bool)

func (x *Scraper) htmlParser(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	switch {
	case x.html1Scraper(siteUrl, body, recipe):
	case x.html2Scraper(siteUrl, body, recipe):
	case x.html3Scraper(siteUrl, body, recipe):
	case x.html4Scraper(siteUrl, body, recipe):
	default:
		return false
	}
	return true
}

func (x *Scraper) html1Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML1RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
				for _, v, more := tokenizer.TagAttr(); more; _, v, more = tokenizer.TagAttr() {
					switch string(v) {
					case "ingredient":
						textIsIngredient = true
						break
					}
				}
			}
		case html.TextToken:
			text := tokenizer.Text()
			if textIsIngredient {
				ingredients = append(ingredients, string(text))
				textIsIngredient = false
			}
		}
	}
	return false
}
func (x *Scraper) html2Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML2RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "li":
				raw := string(tokenizer.Raw())
				if strings.Contains(raw, "ingredient") {
					textIsIngredient = true
				}
			}
		case html.TextToken:
			text := tokenizer.Text()
			if textIsIngredient {
				//fmt.Println(string(text))
				ingredients = append(ingredients, string(text))
				textIsIngredient = false
			}
		}
	}
	return false
}
func (x *Scraper) html3Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML3RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "span":
				raw := string(tokenizer.Raw())
				if strings.Contains(raw, "ingredients") {
					textIsIngredient = true
				}
			}
		case html.TextToken:
			text := tokenizer.Text()
			if textIsIngredient {
				//fmt.Println(string(text))
				ingredients = append(ingredients, string(text))
				textIsIngredient = false
			}
		}
	}
	return false
}
func (x *Scraper) html4Scraper(siteUrl string, body []byte, recipe *RecipeObject) (found bool) {
	textIsIngredient := false
	ingredients := make([]string, 0)
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if len(ingredients) > 0 {
				recipe.Type = HTML4RecipeType
				recipe.RecipeIngredient = ingredients
				return true
			}
			return false
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				raw := string(tokenizer.Raw())
				if strings.Contains(raw, "ingredients") {
					textIsIngredient = true
				}
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			switch string(name) {
			case "div":
				textIsIngredient = false
			}
		case html.TextToken:
			text := string(tokenizer.Text())
			if textIsIngredient{
				text = strings.TrimRight(text, "\n ")
				text = strings.TrimSpace(text)
				if text == "" {
					continue
				}
				if text == "Ingredients" {
					continue
				}
				ingredients = append(ingredients, text)
			}
		}
	}
	return false
}
