package compare

import (
	"fmt"
	"sync"

	"scraper/internal/recipeclient"
	"scraper/internal/util"

	log "github.com/sirupsen/logrus"
)

type Compare struct {
	wg          *sync.WaitGroup
	csvFileName *string
	csvLines    [][]string
	urlIdx      int
	siteClient  *recipeclient.SiteClient
}

func New() *Compare {
	out := &Compare{
		wg:         &sync.WaitGroup{},
		csvLines:   make([][]string, 0),
		siteClient: recipeclient.NewSiteClient(),
	}
	return out
}

func (x *Compare) Main() {
	filename, err := util.CommandParams()
	if err != nil {
		log.Fatal(err)
		return
	}

	c := util.New(*filename)
	if err = c.Open(); err != nil {
		log.Fatal(err)
		return
	}

	legacy := NewLegacy()
	rc := recipeclient.New()

	// loop through all of the urls
	x.csvLines = append(x.csvLines, []string{"url", "scraper", "Legacy", "New"})
	for sourceURL := range c.C {
		// get web page directly
		recipeError := ""
		recipe, err := rc.GetRecipe(sourceURL)
		if err != nil {
			recipeError = err.Error()
		}
		// get legacy scraped information
		legacyData := legacy.Get(sourceURL)
		scraper := ""
		switch {
		case recipe == nil:
		case recipe.Scraper != "":
			scraper = recipe.Scraper
		}
		if recipeError != "" || legacyData.Error != "" {
			x.csvLines = append(x.csvLines, []string{sourceURL, scraper, recipeError, legacyData.Error})
			fmt.Println(sourceURL)
			if legacyData.Error != "" {
				fmt.Println("\tPHP Err:\t", legacyData.Error)
			}
			if recipeError != "" {
				fmt.Println("\tNew Err:\t", recipeError)
			}
			continue
		}

		fmt.Printf("%s\t%s\n", sourceURL, recipe.Scraper)
		for idx, row1 := range legacyData.Data.Items {
			if !searchStringList(recipe.RecipeIngredient, row1.Text) {
				x.csvLines = append(x.csvLines, []string{sourceURL, recipe.Scraper, row1.Text, recipe.RecipeIngredient[idx]})
				fmt.Println("PHP", idx, "\t", row1.Text)
				fmt.Println("________")
				for _, l := range legacyData.Data.Items {
					fmt.Println("Old\t", l.Text)
				}
				fmt.Println("________")
				for _, n := range recipe.RecipeIngredient {
					fmt.Println("New\t", n)
				}
				fmt.Println("URL:", sourceURL)
				fmt.Println("Scraper", recipe.Scraper[0])
				fmt.Println("")
				break
			}
		}
	}
}

func searchStringList(list []string, match string) bool {
	for _, a := range list {
		if a == match {
			return true
		}
	}
	return false
}
