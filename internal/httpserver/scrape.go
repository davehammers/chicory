package httpserver

// supports the /scrape and /scrapeall URL endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scraper/internal/recipeclient"
	"scraper/internal/scraper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type IngredientResponse struct {
	Ingredients []string `json:"ingredients"`
}

type ChicoryResponse struct {
	Data Data `json:"data"`
}
type Items struct {
	Text string `json:"text"`
}
type Data struct {
	Name              string  `json:"name"`
	RecipeCategory   string   `json:"recipeCategory"`
	RecipeCuisine    string   `json:"recipeCuisine"`
	SourceURI         string  `json:"source_uri"`
	ImageURI          string  `json:"image_uri"`
	Items             []Items `json:"items"`
}

// scrape - dump all of the found recipe details to the browser
func (x *Server) scrape(c *fiber.Ctx) (err error) {
	sourceURL := c.OriginalURL()
	parts := strings.SplitN(sourceURL, "=", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("No url parameter specified")
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	sourceURL = parts[1]
	fmt.Println(sourceURL)

	recipe, err := x.client.GetRecipe(sourceURL)
	switch err.(type) {
	case nil:
	case recipeclient.NotFoundError:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	default:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	b, err := JSONMarshal(recipe)
	if err == nil {
		return c.SendString(string(b))
	}
	return c.SendStatus(http.StatusNotFound)
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func (x *Server) getChicoryScrape(c *fiber.Ctx) (err error) {
	sourceURL := c.OriginalURL()
	parts := strings.SplitN(sourceURL, "=", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("No url parameter specified")
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	sourceURL = parts[1]
	fmt.Println(sourceURL)

	recipe, err := x.client.GetRecipe(sourceURL)
	switch err.(type) {
	case nil:
	case recipeclient.NotFoundError:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	default:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	cResp := x.marshalChicoryResponse(recipe)
	b, err := json.MarshalIndent(cResp,"", "    ")
	if err == nil {
		return c.SendString(string(b))
	}
	return c.SendStatus(http.StatusNotFound)
}

func (x *Server) postChicoryScrape(c *fiber.Ctx) (err error) {
	return
}

func (x *Server) marshalChicoryResponse(recipe *scraper.RecipeObject) (cResp ChicoryResponse) {
	cResp.Data.Name = recipe.Name
	cResp.Data.RecipeCuisine = recipe.RecipeCuisine
	cResp.Data.RecipeCategory = recipe.RecipeCategory
	cResp.Data.SourceURI = recipe.SourceURL
	cResp.Data.ImageURI = recipe.Image
	cResp.Data.Items = make([]Items,0)
	for _, row := range recipe.RecipeIngredient {
		cResp.Data.Items = append(cResp.Data.Items, Items{Text: row})
	}
	return
}
