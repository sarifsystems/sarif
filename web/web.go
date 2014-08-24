package web

import (
	"net/http"

	"code.google.com/p/go.net/websocket"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

type Config struct {
	Interface string
}

type Server struct {
	cfg Config
	ctx *core.Context
}

func New(ctx *core.Context) (*Server, error) {
	var cfg Config
	err := ctx.Config.Get("web", &cfg)
	s := &Server{
		cfg,
		ctx,
	}
	return s, err
}

func (s *Server) Start() error {
	http.Handle("/", http.FileServer(http.Dir("assets/web")))
	http.Handle("/stream/stark", websocket.Handler(s.handleStreamStark))

	go func() {
		s.ctx.Log.Infof("[web] listening on %s", s.cfg.Interface)
		err := http.ListenAndServe(s.cfg.Interface, nil)
		s.ctx.Log.Warnln(err)
	}()
	return nil
}

func (s *Server) handleStreamStark(ws *websocket.Conn) {
	defer ws.Close()
	mtp := s.ctx.Proto.NewEndpoint()
	s.ctx.Log.Infoln("[web-socket] new connection")

	webtp := proto.NewByteEndpoint(ws)
	webtp.RegisterHandler(func(msg proto.Message) {
		if err := mtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web-mtp] ", err)
		}
	})
	mtp.RegisterHandler(func(msg proto.Message) {
		if err := webtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web-socket] ", err)
		}
	})
	err := webtp.Listen()
	s.ctx.Log.Errorln("[web-socket] ", err)
}
