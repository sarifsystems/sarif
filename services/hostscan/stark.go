package hostscan

import (
	"database/sql"
	"github.com/xconstruct/stark/client"
	"github.com/xconstruct/stark/log"
)

type Service struct {
	scan *HostScan
}

func NewService(db *sql.DB) *Service {
	s := &Service{
		&HostScan{db},
	}
	return s
}

func (s *Service) Enable(client *client.Client) error {
	return client.Subscribe("devices/fetch_last_status", s.HandleLastStatus)
}

func (s *Service) HandleLastStatus(c *client.Client, msg client.Message) {
	if name := msg.PayloadGetString("host"); name != "" {
		host, err := s.scan.LastStatus(name)
		log.Default.Debugln(host)
		if err != nil {
			log.Default.Warnln(err)
			return
		}
		c.Publish(msg.Reply(client.Message{
			Action: "devices/last_status",
			Payload: map[string]interface{}{
				"host": host,
				"text": host.String(),
			},
		}))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	log.Default.Debugln(hosts)
	if err != nil {
		log.Default.Warnln(err)
		return
	}
	c.Publish(msg.Reply(client.Message{
		Action: "devices/last_status",
		Payload: map[string]interface{}{
			"hosts": hosts,
			"text":  "",
		},
	}))
}
