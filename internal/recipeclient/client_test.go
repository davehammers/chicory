// +build unit

package recipeclient

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

// TestMockGetRecipe - read the HTML from a local file for unit testing
func TestMockGetRecipe(t *testing.T) {

	x := New().
		SetClient(&mockClient{}).
		NewClient()

	list, err := x.GetRecipe("https://anythingwilldo.com")
	assert.Nil(t, err)
	assert.NotEqual(t, len(list), 0, "no recipies found")
	for _, recipe := range list {
		b, _ := json.MarshalIndent(recipe, "", "    ")
		t.Log(string(b))
	}

}

// TestHttpGetRecipe - GET from a URL for integration testing
func TestHttpGetRecipe(t *testing.T) {
	x := New().
		SetClient(&http.Client{Timeout: 20 * time.Second}).
		NewClient()

	list, err := x.GetRecipe("https://www.bettycrocker.com/recipes/lemon-raspberry-bars/5aaa9c08-53f9-404f-89e0-47ef9e49e605")
	assert.Nil(t, err)
	assert.NotEqual(t, len(list), 0, "no recipies found")
	for _, recipe := range list {
		b, _ := json.MarshalIndent(recipe, "", "    ")
		t.Log(string(b))
	}

}

// Do - Mock HTTP interface that reads HTML from a local file
func (x *mockClient) Do(req *http.Request) (resp *http.Response, err error) {
	data, err := os.Open("testdata/site1.html")
	resp = &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       data,
		Request:    req,
	}
	return
}
