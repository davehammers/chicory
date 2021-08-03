// the config package initializes all of the packages used by this application.
// The Start() function is called to start the application once all of the component packages have
// been configured.
package config

import (
	"flag"
	"fmt"
	"scraper/internal/httpserver"
	"scraper/internal/recipeclient"
	"scraper/internal/scraper"
	"scraper/internal/util"
)

type Config struct {
	RecipeClient  *recipeclient.RecipeClient
	Server        *httpserver.Server
	serverAddress string
	serverPort    int
	scrape scraper.Scraper
}

// New - returns an empty *Config
func New() *Config {
	return &Config{}
}

func (x *Config) Main() {
	x.ConfigScraper().
		ConfigRecipeClient().
		CommandLineOptions().
		ConfigServer().
		Start()
}

func (x *Config) CommandLineOptions() *Config {
	flag.StringVar(&x.serverAddress, "a", "0.0.0.0", "Server IP address")
	flag.IntVar(&x.serverPort, "p", 9000, "Server port number")
	flag.Parse()
	return x
}

func (x *Config) ConfigScraper() *Config {
	return x
}

// ConfigRecipeClient - configure and initialize an instance of the recipeclient package
func (x *Config) ConfigRecipeClient() *Config {
	x.RecipeClient = recipeclient.New()
	return x
}

// ConfigServer - configure and initialize an instance of the httpserver  package
func (x *Config) ConfigServer() *Config {
	serverIn := httpserver.New().
		SetAddress(util.LookupEnv("SERVER_ADDRESS", x.serverAddress)).
		SetPort(util.LookupEnvInt("SERVER_PORT", x.serverPort)).
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
