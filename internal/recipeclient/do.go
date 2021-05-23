package recipeclient

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// define an abstract for an HTTP client interface
// This allows for passing in mock interfaces for testing
type RClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client - control data structure for client I/O
type Client struct {
	client RClient
}

// NewHttpClient - allocate a new client data structure
func NewHttpClient() *Client {
	return &Client{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

// Do Sends the *http.Request and returns *http.Response.
// the transaction is retried on error
// There are placeholders for login authentication and rate limiting
func (x *Client) Do(req *http.Request) (resp *http.Response, err error) {
	switch req.Method {
	case http.MethodGet:
		// GET transactions have a rate limit
		// TODO x.getLimiter.Wait()
	default:
		// all other HTTP transaction limit
		// TODO x.updLimiter.Wait()
	}
	for retry := 0; retry < 3; retry++ {
		// pass through request to client
		resp, err = x.client.Do(req)
		if err != nil {
			log.Warn(err)
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			err = nil
			return
		case http.StatusUnauthorized:
			// if our auth has expired, re-auth and try again
			// TODO
		case http.StatusTooManyRequests:
			// throttling feedback from GCP
			// TODO
			log.Warn("HTTP status", resp.StatusCode)
		}
	}
	return
}
