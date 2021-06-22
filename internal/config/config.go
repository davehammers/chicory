// the config package initializes all of the packages used by this application.
// The Start() function is called to start the application once all of the component packages have
// been configured.
package config

import (
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"rscan/internal/httpserver"
	"rscan/internal/recipeclient"
)

type Config struct {
	RecipeClient *recipeclient.RecipeClient
	Server       *httpserver.Server
	Redis        *redis.Client
}

// New - returns an empty *Config
func New() *Config {
	return &Config{}
}

func (x *Config) ConfigRedis() *Config {
	x.Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return x
}

// ConfigRecipeClient - configure and initialize an instance of the recipeclient package
func (x *Config) ConfigRecipeClient() *Config {
	x.RecipeClient = recipeclient.New().
		SetClient(&http.Client{Timeout: 20 * time.Second}).
		SetRedis(x.Redis).
		NewClient()
	return x
}

// ConfigServer - configure and initialize an instance of the httpserver  package
func (x *Config) ConfigServer() *Config {
	x.Server = httpserver.New().
		SetAddress("0.0.0.0").
		SetPort(8080).
		SetClient(x.RecipeClient).
		NewServer()

	x.Server.AddRoutes()
	return x
}

// Start - called after all config is complete, it starts the application running. Start()
// is not expected to return unless the application has failed.
func (x *Config) Start() *Config {
	// this is a blocking function call. It will only return if there is a server error
	x.Server.Handler()
	return x
}
