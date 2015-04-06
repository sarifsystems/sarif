// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/pkg/schema"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Conversation struct {
	Device      string
	LastTime    time.Time
	LastMessage proto.Message

	LastUserText    string
	LastUserMessage proto.Message
}

type Service struct {
	Config services.Config
	Log    proto.Logger
	*proto.Client

	parser        *natural.LearningParser
	conversations map[string]*Conversation
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Config,
		deps.Log,
		deps.Client,

		natural.NewLearningParser(),
		make(map[string]*Conversation),
	}
}

func (s *Service) Enable() error {
	if err := s.Subscribe("natural/handle", "", s.handleNatural); err != nil {
		return err
	}
	if err := s.Subscribe("natural/parse", "", s.handleNaturalParse); err != nil {
		return err
	}
	if err := s.Subscribe("", "user", s.handleUserMessage); err != nil {
		return err
	}

	if err := s.loadModel(); err != nil {
		return err
	}
	go func() {
		time.Sleep(1 * time.Minute)
		if err := s.saveModel(); err != nil {
			s.Log.Errorln("[natural] error saving model:", err)
			return
		}
		c := time.Tick(time.Hour)
		for n := range c {
			if err := s.saveModel(); err != nil {
				s.Log.Errorln("[natural] error saving model:", err, n)
			}
		}
	}()
	return nil
}

func (s *Service) loadModel() error {
	path := s.Config.Dir() + "/" + "natural.json.gz"
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.Log.Debugln("[natural] loading default model")
			natural.TrainDefaults(s.parser)
			return nil
		}
		return err
	}
	defer f.Close()
	s.Log.Debugln("[natural] loading model from", path)
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(gz)

	model := &natural.Model{}
	if err := dec.Decode(model); err != nil {
		return err
	}
	return s.parser.LoadModel(model)
}

func (s *Service) saveModel() error {
	path := s.Config.Dir() + "/" + "natural.json.gz"
	s.Log.Debugln("[natural] saving model to", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	model := s.parser.Model()
	if err := json.NewEncoder(gz).Encode(model); err != nil {
		return err
	}
	return gz.Close()
}

func (s *Service) Disable() error {
	return nil
}

type MsgErrNatural struct {
	Original string `json:"original"`
}

func (pl MsgErrNatural) String() string {
	return "I didn't understand your message."
}

type MsgNaturalParsed struct {
	Parsed   proto.Message
	Original string `json:"original"`
}

func (pl MsgNaturalParsed) String() string {
	return "Natural message correctly parsed."
}

func (s *Service) parseNatural(msg proto.Message) (proto.Message, bool) {
	if msg.Text == "" {
		return proto.Message{}, false
	}

	if parsed, ok := natural.ParseSimple(msg.Text); ok {
		return parsed, ok
	}
	parsed, w := s.parser.Parse(msg.Text)
	s.Log.Debugf(`[natural] parsed "%s" as action "%s" (%g)`, msg.Text, parsed.Action, w)

	if parsed.Text == "" {
		parsed.Text = msg.Text
	}
	return parsed, w > 2
}

func (s *Service) getConversation(device string) *Conversation {
	cv, ok := s.conversations[device]
	if !ok {
		cv = &Conversation{
			Device: device,
		}
		s.conversations[device] = cv
		s.Subscribe("", s.DeviceId+"/"+device, s.handleNetworkMessage)
	}
	return cv
}

type Actionable struct {
	Action  *schema.Action   `json:"action"`
	Actions []*schema.Action `json:"actions"`
}

func (a Actionable) IsAction() bool {
	if a.Action == nil {
		return false
	}
	if a.Action.SchemaType == "TextEntryAction" {
		return true
	}

	return false
}

func answer(a *schema.Action, text string) (proto.Message, bool) {
	reply := proto.Message{
		Action: a.Reply,
		Text:   text,
	}
	if text == ".cancel" || text == "cancel" || strings.HasPrefix(text, "cancel ") {
		return reply, false
	}
	if a.SchemaType == "TextEntryAction" {
		return reply, true
	}
	return reply, false
}

func (s *Service) publishForClient(cv *Conversation, msg proto.Message) {
	msg.Source = s.DeviceId + "/" + cv.Device
	s.Publish(msg)
}

func (s *Service) forwardToClient(cv *Conversation, msg proto.Message) {
	// Save conversation.
	cv.LastTime = time.Now()
	cv.LastMessage = msg

	// Forward response to client.
	msg.Id = proto.GenerateId()
	msg.Destination = cv.Device
	natural.FormatMessage(&msg)
	s.Publish(msg)
}

func (s *Service) handleNatural(msg proto.Message) {
	cv := s.getConversation(msg.Source)

	if msg.Text == ".full" {
		text, err := json.MarshalIndent(cv.LastMessage, "", "    ")
		if err != nil {
			panic(err)
		}
		s.Reply(msg, proto.Message{
			Action: "natural/full",
			Text:   string(text),
		})
		return
	}

	// Check if client answers a conversation.
	if time.Now().Sub(cv.LastTime) < 5*time.Minute {
		var act Actionable
		if err := cv.LastMessage.DecodePayload(&act); err == nil {
			if act.IsAction() {
				parsed, ok := answer(act.Action, msg.Text)
				parsed.Destination = cv.LastMessage.Source
				cv.LastTime = time.Time{}
				if ok {
					s.publishForClient(cv, parsed)
				}
				return
			}
		}
	}

	// Otherwise parse message as normal request.
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}
	cv.LastUserText = msg.Text
	cv.LastUserMessage = parsed
	parsed.CorrId = msg.Id
	s.publishForClient(cv, parsed)
}

func (s *Service) handleNaturalParse(msg proto.Message) {
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}

	reply := proto.CreateMessage("natural/parsed", MsgNaturalParsed{parsed, msg.Text})
	s.Publish(msg.Reply(reply))
}

func (s *Service) handleNetworkMessage(msg proto.Message) {
	client := strings.TrimPrefix(msg.Destination, s.DeviceId+"/")
	cv := s.getConversation(client)
	s.forwardToClient(cv, msg)
}

func (s *Service) handleUserMessage(msg proto.Message) {
	for _, cv := range s.conversations {
		s.forwardToClient(cv, msg)
	}
}
