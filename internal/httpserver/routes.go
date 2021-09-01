package httpserver

// AddRoutes - add the endpoint URLs to the HTTP server
func (x *Server) AddRoutes() {
	x.server.Get("/scrape", x.scrape)
	x.server.Get("/api/scrape/recipe", x.getChicoryScrape)
	x.server.Post("/api/scrape/recipe", x.postChicoryScrape)
}
