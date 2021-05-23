package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

type SchmaRecipe struct {
	Context         string `json:"@context,omitempty"`
	Type            string `json:"@type,omitempty"`
	AggregateRating struct {
		Type        string `json:"@type,omitempty"`
		BestRating  string `json:"bestRating,omitempty"`
		RatingCount string `json:"ratingCount,omitempty"`
		RatingValue string `json:"ratingValue,omitempty"`
		WorstRating string `json:"worstRating,omitempty"`
	} `json:"aggregateRating,omitempty"`
	Author struct {
		Type string `json:"@type,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"author,omitempty"`
	DateCreated string `json:"dateCreated,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Name        string `json:"name,omitempty"`
	Nutrition   struct {
		Type                string `json:"@type,omitempty"`
		Calories            string `json:"calories,omitempty"`
		CarbohydrateContent string `json:"carbohydrateContent,omitempty"`
		CholesterolContent  string `json:"cholesterolContent,omitempty"`
		FatContent          string `json:"fatContent,omitempty"`
		FiberContent        string `json:"fiberContent,omitempty"`
		ProteinContent      string `json:"proteinContent,omitempty"`
		SaturatedFatContent string `json:"saturatedFatContent,omitempty"`
		ServingSize         string `json:"servingSize,omitempty"`
		SodiumContent       string `json:"sodiumContent,omitempty"`
		SugarContent        string `json:"sugarContent,omitempty"`
		TransFatContent     string `json:"transFatContent,omitempty"`
	} `json:"nutrition,omitempty"`
	PrepTime           string   `json:"prepTime,omitempty"`
	RecipeCategory     string   `json:"recipeCategory,omitempty"`
	RecipeIngredient   []string `json:"recipeIngredient,omitempty"`
	RecipeInstructions []struct {
		Type string `json:"@type,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"recipeInstructions,omitempty"`
	RecipeYield string `json:"recipeYield,omitempty"`
	TotalTime   string `json:"totalTime,omitempty"`
}

func main() {
	resp, err := http.Get("https://www.bettycrocker.com/recipes/lemon-raspberry-bars/5aaa9c08-53f9-404f-89e0-47ef9e49e605")
	if err != nil {
		fmt.Println(err)
		return
	}
	doc, err := html.Parse(resp.Body)
	node(doc)
}

func node(n *html.Node) {
	recipe := &SchmaRecipe{}
	err := json.Unmarshal([]byte(n.Data), recipe)
	if err == nil {
		b, _ := json.MarshalIndent(recipe, "", "    ")
		fmt.Println(string(b))
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		node(c)
	}
}
