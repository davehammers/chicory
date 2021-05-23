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
	return x
}

func (x *Config) Start() *Config {
	return x
}
