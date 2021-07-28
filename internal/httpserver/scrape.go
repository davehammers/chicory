package httpserver

// supports the /scrape and /scrapeall URL endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scraper/internal/recipeclient"

	"github.com/gofiber/fiber/v2"
)

type IngredientResponse struct {
	Ingredients []string `json:"ingredients"`
}

// scrape - dump the ingredients for a single recipe to the browser
func (x *Server) scrape(c *fiber.Ctx) (err error) {
	siteURL := c.Query("url")
	fmt.Println(siteURL)

	recipe, err := x.client.GetRecipies(siteURL)
	switch err.(type) {
	case nil:
	case recipeclient.NotFoundError:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	default:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	resp := IngredientResponse{
		Ingredients: recipe.RecipeIngredient,
	}
	b, err := JSONMarshal(resp)
	if err != nil {
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	}
	return c.SendString(string(b))
}

// scrapeAll - dump all of the found recipe details to the browser
func (x *Server) scrapeAll(c *fiber.Ctx) (err error) {
	siteURL := c.Query("url")
	fmt.Println(siteURL)

	recipe, err := x.client.GetRecipies(siteURL)
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
	if err != nil {
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	}
	return c.SendString(string(b))
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}