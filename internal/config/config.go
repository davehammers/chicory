package config

import (
	"net/http"
	"time"

	"rscan/internal/httpserver"
	"rscan/internal/recipeclient"
)

type Config struct {
	RecipeClient *recipeclient.RecipeClient
	Server       *httpserver.Server
}

func New() *Config {
	return &Config{}
}

func (x *Config) ConfigRecipeClient() *Config {
	x.RecipeClient = recipeclient.New().
		SetClient(&http.Client{Timeout: 20 * time.Second}).
		NewClient()
	return x
}

func (x *Config) ConfigServer() *Config {
	x.Server = httpserver.New().
		SetAddress("0.0.0.0").
		SetPort(8080).
		SetClient(x.RecipeClient).
		NewServer()

	x.Server.AddRoutes()
	return x
}

func (x *Config) Start() *Config {
	// this is a blocking function call. It will only return if there is a server error
	x.Server.Handler()
	return x
}
