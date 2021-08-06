// httpserver configures and initializes this applications URLs. The endpoint URLs are registered with the server so it can properly route the transaction to the correct function

package httpserver

// These functions initialize the package for each server instance

import (
	"context"
	"fmt"

	"scraper/internal/recipeclient"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

const (
	defaultHttpPort = 80
)

// HttpServerIn - called during config to pass in parameters needed by the server
type HttpServerIn struct {
	Address string
	Port    int
	Client  *recipeclient.RecipeClient
}

// Server - internal data structures for the HTTP server
type Server struct {
	// router. See github.com/gorilla/mux documentation for full details
	ctx     context.Context
	server  *fiber.App
	address string
	port    int
	client  *recipeclient.RecipeClient
}

// New - returns an empty *HttpServerIn
func New() *HttpServerIn {
	return &HttpServerIn{}
}

// SetAddress - sets the HttpServerIn.Address data field
func (x *HttpServerIn) SetAddress(in string) *HttpServerIn {
	x.Address = in
	return x
}

// SetPort - sets the HttpServerIn.SetPort data field
func (x *HttpServerIn) SetPort(in int) *HttpServerIn {
	x.Port = in
	return x
}

// SetClient - sets the HttpServerIn.Client data field
func (x *HttpServerIn) SetClient(in *recipeclient.RecipeClient) *HttpServerIn {
	x.Client = in
	return x
}

// NewServer - creats a new server data structure instance from the HttpServerIn parameters
func (x *HttpServerIn) NewServer() (out *Server) {
	cfg := fiber.Config{
		BodyLimit:       8 * 1024,
		WriteBufferSize: 8 * 1024,
		GETOnly:         true,
		AppName:         "scraper",
	}
	out = &Server{
		server: fiber.New(cfg),
		client: x.Client,
	}
	// out.server.Use(cors.ConfigDefault)
	out.AddRoutes()

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
Handler starts a CORS enabled HTTP handler listening for inbound HTTP requests.
This is a blocking function and will only return if there is a server failure.

Use
	go x.Handler()
to run this as a background server
*/
func (x *Server) Handler() {
	err := x.server.Listen(fmt.Sprintf("%s:%d", x.address, x.port))
	// This function is blocking. log if this returns
	log.Fatal(err)
}
