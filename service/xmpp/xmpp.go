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
	ReplyTo string
}

type XmppService struct {
	service *service.Service
	client *xmpp.Client
	conf Config
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
	return &XmppService{s, client, conf}, nil
}

func NewService(url string, conf map[string]interface{}) (*XmppService, error) {
	confStruct := Config{}
	confStruct.Host, _ = conf["host"].(string)
	confStruct.User, _ = conf["user"].(string)
	confStruct.Password, _ = conf["password"].(string)
	confStruct.ReplyTo, _ = conf["reply_to"].(string)
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
			x.client.Send(xmpp.Chat{
				Remote: x.conf.ReplyTo,
				Type: "chat",
				Text: msg.Message,
			})
		}
	}()
}
