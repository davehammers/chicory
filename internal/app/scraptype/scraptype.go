package scraptype

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"scraper/internal/recipeclient"

	log "github.com/sirupsen/logrus"
)

// HTTP summary
type bySite map[string]int
type byCode map[int]int
type byScraper map[string]int

type siteCode map[string]byCode
type codeSite map[int]bySite

// by scraper
type siteScraper map[string]byScraper
type scraperSite map[string]bySite

// by URL
type urlScraper map[string]string

type scraperType struct {
	sumCodeSite    codeSite
	sumSiteCode    siteCode
	sumSiteScraper siteScraper
	sumScraperSite scraperSite
	sumURLScraper  urlScraper
	wg             *sync.WaitGroup
	csvFileName    *string
	csvLines       [][]string
	urlIdx         int
	siteClient     *recipeclient.SiteClient
}

func New() *scraperType {
	out := &scraperType{
		sumCodeSite:    make(codeSite),
		sumSiteCode:    make(siteCode),
		sumSiteScraper: make(siteScraper),
		sumScraperSite: make(scraperSite),
		sumURLScraper: make(urlScraper),
		wg:             &sync.WaitGroup{},
		csvLines:       make([][]string,0),
		siteClient:     recipeclient.NewSiteClient(),
	}
	return out
}
func (x *scraperType) Main() {
	if err := x.commandParams(); err != nil {
		log.Fatal(err)
		return
	}
	if err := x.loadCSV(); err != nil {
		log.Fatal(err)
		return
	}

	x.wg.Add(1)
	// start client worker
	go x.worker()

	if err := x.processCSV(); err != nil {
		log.Fatal(err)
		return
	}
	// wait here until all URLs have been processed
	x.wg.Wait()

	// output the summary data
	x.updateCSV()
	fmt.Println(strings.Repeat("-", 80))
	x.printHttpSummary()
	x.printSiteSummary()
	x.printScraperSummary()
}

func (x *scraperType) commandParams() (err error){
	x.csvFileName = flag.String("f", "", "CSV file with urls produced from Superset")
	flag.Parse()
	switch {
	case x.csvFileName == nil:
		err = fmt.Errorf("CSV filename required")
	case *x.csvFileName == "":
		err = fmt.Errorf("CSV filename missing")
	}
	return
}

func (x *scraperType) loadCSV() (err error) {
	f, err := os.Open(*x.csvFileName)
	if err != nil {
		log.Error(err)
		return
	}
	x.csvLines, err = csv.NewReader(f).ReadAll()
	f.Close()
	if len(x.csvLines) == 0 {
		err = fmt.Errorf("No records found in CSV file %s", x.csvFileName)
		return
	}

	// find the URL column
	for idx, val := range x.csvLines[0] {
		if val = strings.ToLower(val);val == "url" {
			x.urlIdx = idx
			break
		}
	}
	return
}

func (x *scraperType) processCSV() (err error){
	for idx, val := range x.csvLines {
		switch idx {
		case 0:
		default:
			siteURL := val[x.urlIdx] // GET the url string
			if siteURL == "" {
				continue
			}
			err = x.siteClient.SiteGetRecipe(siteURL)
			if err != nil {
				log.Println("URL Error", siteURL, err)
			}
		}
	}
	return
}

func (x *scraperType) updateCSV() (err error) {
	for idx, row := range x.csvLines{
		switch idx {
		case 0:
			x.csvLines[0] = append(x.csvLines[0], "scraper")
		default:
			urlStr := row[x.urlIdx]
			// find the url scraper
			if scraper, ok := x.sumURLScraper[urlStr]; ok{
				x.csvLines[idx] = append(x.csvLines[idx], scraper)
			} else {
				x.csvLines[idx] = append(x.csvLines[idx], "")
			}
		}
	}
	f, err := os.Create(*x.csvFileName+".csv")
	csv.NewWriter(f).WriteAll(x.csvLines)
	f.Close()
	return
}

func (x *scraperType) worker() {
	activity := false
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			if !activity {
				x.wg.Done()
				return
			}
			activity = false
		case recipe := <-x.siteClient.ReplyChan:
			// count by HTTP code
			activity = true
			u, _ := url.Parse(recipe.SiteURL)
			// count statusCode/Site
			if _, ok := x.sumCodeSite[recipe.StatusCode]; !ok {
				x.sumCodeSite[recipe.StatusCode] = make(bySite)
			}
			x.sumCodeSite[recipe.StatusCode][u.Host]++

			// count site/StatusCode
			if _, ok := x.sumSiteCode[u.Host]; !ok {
				x.sumSiteCode[u.Host] = make(byCode)
			}
			x.sumSiteCode[u.Host][recipe.StatusCode]++

			switch recipe.StatusCode {
			case http.StatusOK:
				scraper := recipe.Scraper[0]
				x.sumURLScraper[recipe.SiteURL] = scraper
				// count sites per scraper
				if _, ok := x.sumScraperSite[scraper]; !ok {
					x.sumScraperSite[scraper] = make(bySite)
				}
				x.sumScraperSite[scraper][u.Host]++
				// count scrapers per site
				if _, ok := x.sumSiteScraper[u.Host]; !ok {
					x.sumSiteScraper[u.Host] = make(byScraper)
				}
				x.sumSiteScraper[u.Host][scraper]++

				b, err := JSONMarshal(recipe)
				if err == nil {
					log.Println(recipe.StatusCode, recipe.SiteURL)
					fmt.Println(string(b))
				}
			default:
				log.Println(recipe.StatusCode, recipe.SiteURL, recipe.Error)
				x.sumURLScraper[recipe.SiteURL] = fmt.Sprintf("HTTP %d",recipe.StatusCode)
			}
		}
	}
}

func (x *scraperType) printHttpSummary() {
	type sortType struct {
		code  int
		count int
	}
	type sortStringType struct {
		text  string
		count int
	}
	typeList := make([]sortType, 0)
	// HTTP Summary
	for k, v := range x.sumCodeSite {
		count := 0
		// add up all of the sites
		for _, v1 := range v {
			count += v1
		}
		typeList = append(typeList, sortType{k, count})
	}
	sort.Slice(typeList, func(i, j int) bool { return typeList[i].count > typeList[j].count })

	// print detail
	fmt.Println("")
	fmt.Println("HTTP Detail")
	for _, t := range typeList {
		fmt.Printf("Count:%6d, HTTP: %d\n", t.count, t.code)
		detailList := make([]sortStringType, 0)
		for k, v := range x.sumCodeSite[t.code] {
			detailList = append(detailList, sortStringType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Site: %s\n", d.count, d.text)
		}
	}
	total := 0
	fmt.Println("")
	fmt.Println("HTTP Summary")
	for _, t := range typeList {
		total += t.count
		fmt.Printf("Count:%6d, HTTP: %d\n", t.count, t.code)
	}
	fmt.Printf("Total:%6d, HTTP\n", total)

}

func (x *scraperType) printSiteSummary() {
	type sortType struct {
		text  string
		count int
	}
	textList := make([]sortType, 0)
	total := 0
	for k, v := range x.sumSiteScraper {
		count := 0
		for _, v1 := range v {
			count += v1
		}
		textList = append(textList, sortType{k, count})
	}
	sort.Slice(textList, func(i, j int) bool { return textList[i].count > textList[j].count })

	fmt.Println("")
	fmt.Println("Site Detail")
	// print detail
	for _, t := range textList {
		fmt.Printf("Count:%6d, Site: %s\n", t.count, t.text)
		detailList := make([]sortType, 0)
		for k, v := range x.sumSiteScraper[t.text] {
			detailList = append(detailList, sortType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Scraper: %s\n", d.count, d.text)
		}
	}
	total = 0
	fmt.Println("")
	fmt.Println("Site Summary")
	for _, t := range textList {
		total += t.count
		fmt.Printf("Count:%6d, Site: %s\n", t.count, t.text)
	}
	fmt.Printf("Total:%6d, Site\n", total)
}
func (x *scraperType) printScraperSummary() {
	type sortType struct {
		text  string
		count int
	}
	textList := make([]sortType, 0)
	total := 0
	for k, v := range x.sumScraperSite {
		count := 0
		for _, v1 := range v {
			count += v1
		}
		textList = append(textList, sortType{k, count})
	}
	sort.Slice(textList, func(i, j int) bool { return textList[i].count > textList[j].count })

	fmt.Println("")
	fmt.Println("Scraper Detail")
	// print detail
	for _, t := range textList {
		fmt.Printf("Count:%6d, Scraper: %s\n", t.count, t.text)
		detailList := make([]sortType, 0)
		for k, v := range x.sumScraperSite[t.text] {
			detailList = append(detailList, sortType{k, v})
		}
		sort.Slice(detailList, func(i, j int) bool { return detailList[i].count > detailList[j].count })
		for _, d := range detailList {
			fmt.Printf("\tCount:%6d, Site: %s\n", d.count, d.text)
		}
	}
	// calculate total
	total = 0
	for _, t := range textList {
		total += t.count
	}
	fmt.Println("")
	fmt.Println("Scraper Summary")
	for _, t := range textList {
		fmt.Printf("Count:%6d, %.1f%% Scraper: %s\n", t.count, float32(t.count)/float32(total)*100.0, t.text)
	}
	fmt.Printf("Total:%6d, Scraper\n", total)
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
