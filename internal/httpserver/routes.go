package httpserver

func (x *Server) AddRoutes() {
	x.router.Methods("GET").Path("/scrape").Queries("url", "{url:.*}").HandlerFunc(x.scrape)
	x.router.Methods("GET").Path("/scrapeall").Queries("url", "{url:.*}").HandlerFunc(x.scrapeAll)
}
