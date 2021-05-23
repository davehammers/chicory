package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type SingleResponse struct {
	Ingredients []string `json:"ingredients"`
}

// scrape - dump the ingredients for a single recipe to the browser
func (x *Server) scrape(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteURL := vars["url"]

	list, err := x.client.GetRecipies(siteURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(list) == 0 {
		http.Error(w, "No recipies found", http.StatusNotFound)
		return
	}
	for _, l := range list {
		resp := SingleResponse{
			Ingredients: l.RecipeIngredient,
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		enc.Encode(resp)
		break
	}
}

// scrapeAll - dump all of the found recipe details to the browser
func (x *Server) scrapeAll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteURL := vars["url"]

	list, err := x.client.GetRecipies(siteURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(list) == 0 {
		http.Error(w, "No recipies found", http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(list)
}
