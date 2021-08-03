package httpserver

// supports the /scrape and /scrapeall URL endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scraper/internal/recipeclient"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type IngredientResponse struct {
	Ingredients []string `json:"ingredients"`
}

// scrape - dump all of the found recipe details to the browser
func (x *Server) scrape(c *fiber.Ctx) (err error) {
	siteURL := c.OriginalURL()
	parts := strings.SplitN(siteURL,"=", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("No url parameter specified")
		c.SendString(err.Error())
		return c.SendStatus(http.StatusBadRequest)
	}
	siteURL = parts[1]
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
