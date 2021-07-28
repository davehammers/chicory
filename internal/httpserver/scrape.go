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

// scrape - dump all of the found recipe details to the browser
func (x *Server) scrape(c *fiber.Ctx) (err error) {
	siteURL := c.Query("url")
	fmt.Println(siteURL)

	recipe, err := x.client.GetRecipe(siteURL)
	switch err.(type) {
	case nil:
	case recipeclient.NotFoundError:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusNotFound)
	default:
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	switch r := recipe.(type) {
	case nil:
		c.SendString("NotFound\n")
		return c.SendStatus(http.StatusNotFound)
	case recipeclient.RecipeSchema1:
		b, err := JSONMarshal(r)
		if err == nil {
			return c.SendString(string(b))
		}
	case recipeclient.RecipeSchema2:
		for _, entry := range r.Graph {
			if len(entry.RecipeIngredient) > 0 {
				b, err := JSONMarshal(entry)
				if err == nil {
					return c.SendString(string(b))
				}
			}
		}
	default:
		b, err := JSONMarshal(r)
		if err == nil {
			return c.SendString(string(b))
		}
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
