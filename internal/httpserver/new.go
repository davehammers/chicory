package httpserver

type HttpServerIn struct{}

type Server struct{}

func New() *HttpServerIn {
	return &HttpServerIn{}
}

func (x *HttpServerIn) NewServer() *Server {
	return &Server{}
}
