package zerone

type Server struct {
}

func NewServer(name string) *Server {
	return &Server{}
}

func (s *Server) Register(rcvr interface{}) error {
	return nil
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return nil
}

func (s *Server) ListenAndServe(network, address string) error {
	return nil
}
