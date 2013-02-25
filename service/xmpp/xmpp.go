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
	*service.Service
	client *xmpp.Client
	conf Config
	lastRemote string
}

func New(conf Config) *XmppService {
	s := service.New(service.Info{
		Name: "xmpp",
	})
	x := &XmppService{s, nil, conf, ""}
	s.Handler = x
	return x
}

func NewService(conf map[string]interface{}) *XmppService {
	confStruct := Config{}
	confStruct.Host, _ = conf["host"].(string)
	confStruct.User, _ = conf["user"].(string)
	confStruct.Password, _ = conf["password"].(string)
	return New(confStruct)
}

func (x *XmppService) Serve() error {
	client, err := xmpp.NewClient(x.conf.Host, x.conf.User, x.conf.Password)
	if err != nil {
		return err
	}
	x.client = client

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
				err := x.Write(msg)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	return x.Service.Serve()
}

func (x *XmppService) Handle(msg *stark.Message) (*stark.Message, error) {
	if x.lastRemote == "" {
		return nil, nil
	}
	x.client.Send(xmpp.Chat{
		Remote: x.lastRemote,
		Type: "chat",
		Text: msg.Message,
	})
	return nil, nil
}
