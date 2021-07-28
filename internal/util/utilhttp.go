// util package contains common/convenient routines to aid in development and debugging
package util

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"path"
	"runtime"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

type UIErrorType struct {
	Errors []UIErrors `json:"errors"`
}

// This is the common error reporting format for all HTTP error reporting
type UIErrors struct {
	ErrorMessage string `json:"errorMessage"`
	Resource     string `json:"resource,omitempty"`
	ErrorCode    int    `json:"errorCode"`
	ReasonCode   int    `json:"reasonCode,omitempty"`
}

// UIMsgType a common format for reporting messages such as JSON logging
type UIMsgType struct {
	// list of messages
	UIMsgs []UIMsg `json:"messages"`
}

// This is the common messaging format for simple HTTP message reporting
type UIMsg struct {
	// any text message
	Message string `json:"message"`

	// a string identifying some internal resource
	Resource string `json:"resource,omitempty"`

	// HTTP status code
	Code int `json:"code"`
}

// DumpJSON() pretty prints any data structure into formatted JSON.
// Set LOG_DEBUG=true for output
func DumpJSON(n interface{}) {

	_, fname, lineno, _ := runtime.Caller(1)
	fname = path.Base(fname)

	j, err := json.MarshalIndent(n, "", "    ")
	if err != nil {
		log.Debugf("%s:%d %+v\n", fname, lineno, err)
		log.Debugf("%s:%d %+v\n", fname, lineno, n)
	} else {
		log.Debugf("%s:%d %+v\n", fname, lineno, string(j))
	}
}

// DumpReqURL - pretty print a requests method and URL
// Set LOG_DEBUG=true for output
func DumpReqURL(r *http.Request) {
	_, fname, lineno, _ := runtime.Caller(1)
	fname = path.Base(fname)

	log.Debugf("%s:%d\n %s %s\n", fname, lineno, r.Method, r.URL.String())
}

// DumpRequest - prints a formatted http.Request
// Set LOG_DEBUG=true for output
func DumpRequest(r *http.Request) {
	_, fname, lineno, _ := runtime.Caller(1)
	fname = path.Base(fname)

	x, _ := httputil.DumpRequest(r, true)
	log.Debugf("%s:%d\n%s\n", fname, lineno, string(x))
}

// DumpResponse - prints a formatted http.Response
// Set LOG_DEBUG=true for output
func DumpResponse(resp *http.Response) {
	_, fname, lineno, _ := runtime.Caller(1)
	fname = path.Base(fname)

	x, _ := httputil.DumpResponse(resp, true)
	log.Debugf("%s:%d\n%s", fname, lineno, string(x))
}

// UIError - used for creating UI error responses by producing a standardized error reporting format
// It is a direct replacement for http.Error(w, <string>, code)
func UIError(w http.ResponseWriter, msg string, code int) {
	uiErr := UIErrorType{Errors: []UIErrors{
		UIErrors{
			ErrorCode:    code,
			ErrorMessage: msg,
			Resource:     "",
		},
	},
	}
	DumpJSON(uiErr)
	j, _ := json.Marshal(&uiErr)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)
}

// UIMessage - used for creating simple UI responses by producing a standardized JSON reporting format
// It is a direct replacement for http.Error(w, <string>, code)
func UIMessage(w http.ResponseWriter, msg string, code int) {
	uiMsg := UIMsgType{
		UIMsgs: []UIMsg{
			UIMsg{
				Code:    code,
				Message: msg,
			},
		},
	}
	DumpJSON(uiMsg)
	j, _ := json.Marshal(&uiMsg)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)
}

// SendGzipJson - Format an HTTP response by encoding JSON into a gzipped payload
// the HTTP header is set to Content-Encoding: gzip
func SendGzipJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	g := gzip.NewWriter(w)
	encoder := json.NewEncoder(g)
	encoder.SetIndent("", "    ")
	encoder.Encode(data)
	g.Close()
}

// PrintRoutes utility to write the list of routes to an output
// Useful for creating documentation
func PrintRoutes(w io.Writer, router *mux.Router) {
	fmt.Fprintln(w, "\nURLs")
	fmt.Fprintln(w, "----")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		queries, _ := route.GetQueriesTemplates()
		for _, method := range methods {
			fmt.Fprintf(w, "%-6s %s", method, path)
			for _, q := range queries {
				fmt.Fprintf(w, "?%-6s", q)
			}
		}
		fmt.Fprintf(w, "\n")
		return nil
	})
}
