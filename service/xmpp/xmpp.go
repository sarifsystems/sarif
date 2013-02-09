package xmpp

import (
	"log"

	"github.com/mattn/go-xmpp"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
)

type Config struct {
	Host string
	User string
	Password string
}

type XmppService struct {
	service *service.Service
	client *xmpp.Client
	conf Config
	lastRemote string
}

func NewXmppService(url string, conf Config) (*XmppService, error) {
	s, err := service.Connect(url, service.Info{
		Name: "xmpp",
	})
	if err != nil {
		return nil, err
	}
	client, err := xmpp.NewClient(conf.Host, conf.User, conf.Password)
	if err != nil {
		return nil, err
	}
	return &XmppService{s, client, conf, ""}, nil
}

func NewService(url string, conf map[string]interface{}) (*XmppService, error) {
	confStruct := Config{}
	confStruct.Host, _ = conf["host"].(string)
	confStruct.User, _ = conf["user"].(string)
	confStruct.Password, _ = conf["password"].(string)
	return NewXmppService(url, confStruct)
}

func (x *XmppService) Start() {
	go func() {
		for {
			event, err := x.client.Recv()
			if err != nil {
				log.Fatal(err)
			}

			if chat, ok := event.(xmpp.Chat); ok {
				if chat.Text == "" {
					continue
				}
				msg := stark.NewMessage()
				msg.Action = "natural.process"
				msg.Message = chat.Text
				x.lastRemote = chat.Remote
				err := x.service.Write(msg)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	go func() {
		for {
			msg, err := x.service.Read()
			if err != nil {
				return
			}
			if x.lastRemote == "" {
				continue
			}
			x.client.Send(xmpp.Chat{
				Remote: x.lastRemote,
				Type: "chat",
				Text: msg.Message,
			})
		}
	}()
}
