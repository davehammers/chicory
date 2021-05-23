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

func TestMockGetRecipies(t *testing.T) {

	x := New().
		SetClient(&mockClient{}).
		NewClient()

	list, err := x.GetRecipies("https://anythingwilldo.com")
	assert.Nil(t, err)
	assert.NotEqual(t, len(list), 0, "no recipies found")
	for _, recipe := range list {
		b, _ := json.MarshalIndent(recipe, "", "    ")
		t.Log(string(b))
	}

}

func TestHttpGetRecipies(t *testing.T) {
	x := New().
		SetClient(&http.Client{Timeout: 20 * time.Second}).
		NewClient()

	list, err := x.GetRecipies("https://www.bettycrocker.com/recipes/lemon-raspberry-bars/5aaa9c08-53f9-404f-89e0-47ef9e49e605")
	assert.Nil(t, err)
	assert.NotEqual(t, len(list), 0, "no recipies found")
	for _, recipe := range list {
		b, _ := json.MarshalIndent(recipe, "", "    ")
		t.Log(string(b))
	}

}

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
