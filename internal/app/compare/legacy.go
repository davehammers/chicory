package compare

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Legacy struct {
	Data struct {
		Name              string `json:"name"`
		Servings          string `json:"servings"`
		Description       string `json:"description"`
		Author            string `json:"author"`
		SourceURI         string `json:"source_uri"`
		ImageURI          string `json:"image_uri"`
		PrepTimeInMinutes int    `json:"prep_time_in_minutes"`
		CookTimeInMinutes int    `json:"cook_time_in_minutes"`
		Source            string `json:"source"`
		Items             []struct {
			Text string `json:"text"`
		} `json:"items"`
	} `json:"data"`
	StatusCode int
	Error      string
}

type legacyClient struct {
	client *http.Client
}

func NewLegacy() *legacyClient {
	out := &legacyClient{
		client: &http.Client{},
	}
	return out
}

func (x *legacyClient) Get(sourceURL string) (out Legacy) {
	type postURL struct {
		URLStr string `json:"url"`
	}
	b, _ := json.Marshal(postURL{URLStr: sourceURL})
	resp, err := x.client.Post("https://prod-scraper.chicoryapp.com/api/scrape/recipe", "application/json", bytes.NewReader(b))
	if err != nil {
		out.Error = err.Error()
		out.StatusCode = http.StatusBadRequest
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		out.Error = err.Error()
		out.StatusCode = http.StatusBadRequest
		return
	}
	resp.Body.Close()
	err = json.Unmarshal(body, &out)
	if err != nil {
		//out.Error = string(body)
		out.Error = err.Error()
		out.StatusCode = http.StatusBadRequest
	}
	return
}
