package http

import (
	"encoding/json"
	"net/http"
	neturl "net/url"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

func init() {
	//transport.Register("http", Connect)
}

type serveConn struct {
	dec *json.Decoder
	enc *json.Encoder
	r *http.Request
	w http.ResponseWriter
}

func (c *serveConn) Read() (*stark.Message, error) {
	msg := stark.NewMessage()
	if err := c.dec.Decode(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *serveConn) Write(msg *stark.Message) error {
	return c.enc.Encode(msg)
}

func (c *serveConn) Close() error {
	return nil
}

type HttpTransport struct {
	address string
	man transport.ConnManager
	server *http.Server
}

func NewHttpTransport(man transport.ConnManager, url string) (*HttpTransport, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	return &HttpTransport{u.Host, man, nil}, nil
}

func (t *HttpTransport) Start() error {
	t.server = &http.Server{Addr: t.address, Handler: t}
	return t.server.ListenAndServe()
}

func (t *HttpTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sc := &serveConn{
		json.NewDecoder(r.Body),
		json.NewEncoder(w),
		r,
		w,
	}
	t.man.Connect(sc)
}

type clientConn struct {
	url string
	client *http.Client
	dec *json.Decoder
	enc *json.Encoder
}

func Connect(url string) (transport.Conn, error) {
	return &clientConn{
		url,
		&http.Client{},
		nil, nil,
	}, nil
}
