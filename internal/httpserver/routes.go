package httpserver

// AddRoutes - add the endpoint URLs to the HTTP server
func (x *Server) AddRoutes() {
	x.server.Get("/scrape", x.scrape)
	x.server.Get("/scrapeall", x.scrapeAll)
}
