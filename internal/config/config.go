// the config package initializes all of the packages used by this application.
// The Start() function is called to start the application once all of the component packages have
// been configured.
package config

import (
	"fmt"
	"net/http"
	"time"

	"rscan/internal/httpserver"
	"rscan/internal/recipeclient"
)

type Config struct {
	RecipeClient *recipeclient.RecipeClient
	Server       *httpserver.Server
}

// New - returns an empty *Config
func New() *Config {
	return &Config{}
}

// ConfigRecipeClient - configure and initialize an instance of the recipeclient package
func (x *Config) ConfigRecipeClient() *Config {
	x.RecipeClient = recipeclient.New().
		SetClient(&http.Client{Timeout: 20 * time.Second}).
		NewClient()
	return x
}

// ConfigServer - configure and initialize an instance of the httpserver  package
func (x *Config) ConfigServer() *Config {
	serverIn := httpserver.New().
		SetAddress("0.0.0.0").
		SetPort(8080).
		SetClient(x.RecipeClient)
	x.Server = serverIn.NewServer()

	x.Server.AddRoutes()
	fmt.Printf("Server configured for %s:%d\n", serverIn.Address, serverIn.Port)
	return x
}

// Start - called after all config is complete, it starts the application running. Start()
// is not expected to return unless the application has failed.
func (x *Config) Start() *Config {
	// this is a blocking function call. It will only return if there is a server error
	fmt.Println("Starting application")
	x.Server.Handler()
	return x
}
