package httpserver

import (
	"fmt"
	"net/http"

	"rscan/internal/recipeclient"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	defaultHttpPort = 80
)

type HttpServerIn struct {
	Address string
	Port    int
	Client  *recipeclient.RecipeClient
}

type Server struct {
	// router. See github.com/gorilla/mux documentation for full details
	router  *mux.Router
	address string
	port    int
	client  *recipeclient.RecipeClient
}

func New() *HttpServerIn {
	return &HttpServerIn{}
}

func (x *HttpServerIn) SetAddress(in string) *HttpServerIn {
	x.Address = in
	return x
}

func (x *HttpServerIn) SetPort(in int) *HttpServerIn {
	x.Port = in
	return x
}

func (x *HttpServerIn) SetClient(in *recipeclient.RecipeClient) *HttpServerIn {
	x.Client = in
	return x
}

func (x *HttpServerIn) NewServer() (out *Server) {
	out = &Server{
		router: mux.NewRouter(),
		client: x.Client,
	}
	if x.Port == 0 {
		out.port = defaultHttpPort
	} else {
		out.port = x.Port
	}

	if x.Address == "" {
		out.address = "0.0.0.0"
	} else {
		out.address = x.Address
	}
	return
}

/*
Router returns the mux router used to register functions with the HTTP server.
See https://pkg.go.dev/github.com/gorilla/mux?tab=doc for the complete router documentation
*/
func (x *Server) Router() *mux.Router {
	return x.router
}

/*
Handler starts a CORS enabled HTTP handler listening for inbound HTTP requests.
This is a blocking function and will only return if there is a server failure.

Use
	go x.Handler()
to run this as a background server
*/
func (x *Server) Handler() {
	headersOk := handlers.AllowedHeaders([]string{
		"*",
		"Authorization",
		"X-Requested-With",
		"Content-Type",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{
		"GET",
		"HEAD",
		"PATCH",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS"})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", x.address, x.port),
		handlers.CORS(headersOk, originsOk, methodsOk)(x.router))
	// This function is blocking. log if this returns
	log.Fatal(err)
}
