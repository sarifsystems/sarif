// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service natural provides a conversational interface to the stark network.
package natural

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	Address   string
	ModelPath string
}

type Dependencies struct {
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Config services.Config
	Cfg    Config
	Log    proto.Logger
	*proto.Client

	parser        *natural.Parser
	phrases       *natural.Phrasebook
	conversations map[string]*Conversation
	rand          *rand.Rand
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Config,
		Config{},
		deps.Log,
		deps.Client,

		natural.NewParser(),
		natural.NewPhrasebook(),
		make(map[string]*Conversation),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) Enable() error {
	s.Subscribe("natural/handle", "", s.handleNatural)
	s.Subscribe("natural/parse", "", s.handleNaturalParse)
	s.Subscribe("natural/learn", "", s.handleNaturalLearn)
	s.Subscribe("natural/reinforce", "", s.handleNaturalReinforce)
	s.Subscribe("natural/phrases", "", s.handleNaturalPhrases)
	s.Subscribe("", "user", s.handleUserMessage)

	s.Cfg.Address = "sir"
	s.Cfg.ModelPath = s.Config.Dir() + "/" + "natural.json.gz"
	s.Config.Get(&s.Cfg)

	if err := s.loadModel(); err != nil {
		return err
	}
	go s.saveModelLoop()
	return nil
}

func (s *Service) saveModelLoop() {
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
}

func (s *Service) loadModel() error {
	s.Log.Debugln("[natural] loading model from", s.Cfg.ModelPath)
	return s.parser.LoadModel(s.Cfg.ModelPath)
}

func (s *Service) saveModel() error {
	s.Log.Debugln("[natural] saving model to", s.Cfg.ModelPath)
	return s.parser.SaveModel(s.Cfg.ModelPath)
}

func (s *Service) Disable() error {
	return nil
}

type MsgNaturalParsed struct {
	*natural.ParseResult
}

func (p MsgNaturalParsed) String() string {
	return p.ParseResult.String()
}

func (s *Service) getConversation(device string) *Conversation {
	cv, ok := s.conversations[device]
	if !ok {
		cv = &Conversation{
			service: s,
			Device:  device,
		}
		s.conversations[device] = cv
		s.Subscribe("", s.DeviceId+"/"+device, s.handleNetworkMessage)
	}
	return cv
}

func (s *Service) handleNatural(msg proto.Message) {
	cv := s.getConversation(msg.Source)
	cv.HandleClientMessage(msg)
}

func (s *Service) handleNaturalParse(msg proto.Message) {
	res, err := s.parser.Parse(msg.Text, &natural.Context{
		Sender:    "user",
		Recipient: "stark",
	})
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	action := "natural/parsed"
	if res.Prediction == nil || res.Prediction.Weight <= 0 {
		action = "err/natural/parsed"
	}

	s.Reply(msg, proto.CreateMessage(action, MsgNaturalParsed{res}))
}

func (s *Service) handleNetworkMessage(msg proto.Message) {
	client := strings.TrimPrefix(msg.Destination, s.DeviceId+"/")
	cv := s.getConversation(client)
	cv.SendToClient(msg)
}

func (s *Service) handleUserMessage(msg proto.Message) {
	for _, cv := range s.conversations {
		cv.SendToClient(msg)
	}
}

type msgLearn struct {
	Sentence string `json:"sentence"`
	Action   string `json:"action"`
}

func (p msgLearn) Text() string {
	return "I learned to associate '" + p.Sentence + "' with " + p.Action + "."
}

func (s *Service) handleNaturalLearn(msg proto.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	p.Sentence = strings.TrimSpace(p.Sentence)
	if p.Sentence == "" {
		return
	}
	p.Action = strings.TrimLeft(p.Action, ".")

	if err := s.parser.LearnRule(p.Sentence, p.Action); err != nil {
		s.ReplyBadRequest(msg, err)
	}
	s.Log.Infof("[natural] associating '%s' with %s", p.Sentence, p.Action)
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", p))
}

func (s *Service) handleNaturalReinforce(msg proto.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Sentence == "" || p.Action == "" {
		s.ReplyBadRequest(msg, errors.New("No sentence or action given."))
		return
	}

	s.Log.Infof("[natural] reinforcing: '%s' with %s", p.Sentence, p.Action)
	s.parser.ReinforceSentence(p.Sentence, p.Action)

	parsed, _ := s.parser.Parse(p.Sentence, &natural.Context{})
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", &msgLearn{p.Sentence, parsed.Prediction.Message.Action}))
}

func (s *Service) handleNaturalPhrases(msg proto.Message) {
	ctx := strings.Trim(strings.TrimPrefix(msg.Action, "natural/phrases"), "/")
	if ctx == "" {
		s.ReplyBadRequest(msg, errors.New("The context seems to be missing."))
		return
	}

	text := s.phrases.Get(ctx)
	s.Reply(msg, proto.Message{
		Action: "natural/phrase",
		Text:   text,
	})
}

func (s *Service) TransformReply(text string) string {
	if s.Cfg.Address == "" || text == "" {
		return text
	}
	if strings.LastIndexAny(text, ".!?") != len(text)-1 {
		return text
	}
	if strings.Contains(text, "\n") || len(text) > 80 {
		return text
	}

	if s.rand.Float32() <= 0.25 {
		text = text[0:len(text)-1] + ", " + s.Cfg.Address + text[len(text)-1:]
	}

	return text
}
