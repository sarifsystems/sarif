package hostscan

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/client"
)

type Service struct {
	scan   *HostScan
	client *client.Client
	ctx    *core.Context
}

func NewService(ctx *core.Context) (*Service, error) {
	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}

	client, err := ctx.Client()
	if err != nil {
		return nil, err
	}

	SetupSchema(db.Driver(), db.DB)

	s := &Service{
		New(db.DB),
		client,
		ctx,
	}
	return s, nil
}

func (s *Service) Enable() error {
	return s.client.Subscribe("devices/fetch_last_status", s.HandleLastStatus)
}

func (s *Service) HandleLastStatus(msg proto.Message) {
	if name := msg.PayloadGetString("host"); name != "" {
		host, err := s.scan.LastStatus(name)
		s.ctx.Log.Debugln(host)
		if err != nil {
			s.ctx.Log.Warnln(err)
			return
		}
		s.client.Publish(msg.Reply(proto.Message{
			Action: "devices/last_status",
			Payload: map[string]interface{}{
				"host": host,
				"text": host.String(),
			},
		}))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	s.ctx.Log.Debugln(hosts)
	if err != nil {
		s.ctx.Log.Warnln(err)
		return
	}
	s.client.Publish(msg.Reply(proto.Message{
		Action: "devices/last_status",
		Payload: map[string]interface{}{
			"hosts": hosts,
			"text":  "",
		},
	}))
}
