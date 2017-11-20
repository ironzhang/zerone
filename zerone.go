package zerone

type Options struct {
	Namespace string
	EtcdURLs  []string
}

type Zerone struct {
}

func New(opts Options) *Zerone {
	return &Zerone{}
}

func (p *Zerone) NewClient(service string) *Client {
	return nil
}

func (p *Zerone) NewServer(service string) *Server {
	return nil
}
