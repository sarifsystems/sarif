// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service natural provides a conversational interface to the sarif network.
package natural

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	Address    string
	Rules      natural.SentenceRuleSet
	Parsers    map[string]*Parser
	Annotators map[string]*Annotator
}

type Dependencies struct {
	Config services.Config
	Client *sarif.Client
}

type Service struct {
	Config services.Config
	Cfg    Config
	*sarif.Client

	ParserKeepAlive time.Duration

	regular       *natural.RegularParser
	phrases       *natural.Phrasebook
	conversations map[string]*Conversation
	rand          *rand.Rand
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		Config: deps.Config,
		Cfg:    Config{},
		Client: deps.Client,

		ParserKeepAlive: 30 * time.Minute,

		regular:       natural.NewRegularParser(),
		phrases:       natural.NewPhrasebook(),
		conversations: make(map[string]*Conversation),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type Parser struct {
	Weight   float64
	LastSeen time.Time `json:"-"`
}

type Annotator struct {
	Enabled  bool
	LastSeen time.Time `json:"-"`
}

func (s *Service) Enable() error {
	s.Subscribe("natural/handle", "", s.handleNatural)
	s.Subscribe("natural/parse", "", s.handleNaturalParse)
	s.Subscribe("natural/learn", "", s.handleNaturalLearn)
	s.Subscribe("natural/phrases", "", s.handleNaturalPhrases)
	s.Subscribe("", "user", s.handleUserMessage)

	s.Cfg.Address = "sir"
	s.Cfg.Rules = natural.DefaultRules
	s.Cfg.Parsers = make(map[string]*Parser)
	s.Cfg.Annotators = make(map[string]*Annotator)
	s.Config.Get(&s.Cfg)

	if err := s.regular.Load(s.Cfg.Rules); err != nil {
		return err
	}

	go func() {
		time.Sleep(5 * time.Second)
		for {
			parsers := s.Discover("natural/parse")
			for msg := range parsers {
				if msg.Source == s.DeviceId {
					continue
				}
				p, ok := s.Cfg.Parsers[msg.Source]
				if !ok {
					p = &Parser{Weight: 0.5}
					s.Cfg.Parsers[msg.Source] = p
					s.Config.Set(&s.Cfg)
				}
				p.LastSeen = time.Now()
			}

			anns := s.Discover("natural/annotate")
			for msg := range anns {
				a, ok := s.Cfg.Annotators[msg.Source]
				if !ok {
					a = &Annotator{Enabled: true}
					s.Cfg.Annotators[msg.Source] = a
					s.Config.Set(&s.Cfg)
				}
				a.LastSeen = time.Now()
			}
			time.Sleep(s.ParserKeepAlive / 2)
		}
	}()
	return nil
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
		cv.PublishForClient(sarif.CreateMessage("natural/client/new", nil))
	}
	return cv
}

func (s *Service) handleNatural(msg sarif.Message) {
	cv := s.getConversation(msg.Source)
	cv.HandleClientMessage(msg)
}

func (s *Service) handleNaturalParse(msg sarif.Message) {
	res, err := s.Parse(&natural.Context{
		Text:      msg.Text,
		Sender:    "user",
		Recipient: "sarif",
	})
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	action := "natural/parsed"
	if len(res.Intents) == 0 || res.Intents[0].Weight <= 0 {
		action = "err/natural/parsed"
	}

	s.Reply(msg, sarif.CreateMessage(action, res))
}

func (s *Service) handleNetworkMessage(msg sarif.Message) {
	client := strings.TrimPrefix(msg.Destination, s.DeviceId+"/")
	cv := s.getConversation(client)
	cv.SendToClient(msg)
}

func (s *Service) handleUserMessage(msg sarif.Message) {
	for _, cv := range s.conversations {
		cv.SendToClient(msg)
	}
}

func (s *Service) handleNaturalPhrases(msg sarif.Message) {
	ctx := strings.Trim(strings.TrimPrefix(msg.Action, "natural/phrases"), "/")
	if ctx == "" {
		s.ReplyBadRequest(msg, errors.New("The context seems to be missing."))
		return
	}

	text := s.phrases.Get(ctx)
	s.Reply(msg, sarif.Message{
		Action: "natural/phrase",
		Text:   text,
	})
}

func (s *Service) AnnotateReply(msg sarif.Message) sarif.Message {
	mu := sync.Mutex{}
	fin := make(chan bool)
	go func() {
		for name, a := range s.Cfg.Annotators {
			if a.Enabled && (time.Now().Sub(a.LastSeen) < s.ParserKeepAlive) {
				reply, ok := <-s.Request(sarif.Message{
					Action:      "natural/annotate/" + msg.Action,
					Destination: name,
					Text:        msg.Text,
					Payload:     msg.Payload,
				})
				if !ok || reply.IsAction("err") {
					continue
				}
				mu.Lock()
				if reply.Text != "" {
					msg.Text = reply.Text
				}
				if len(reply.Payload.Raw) > 0 {
					msg.Payload = reply.Payload
				}
				mu.Unlock()
			}
		}
		fin <- true
	}()

	select {
	case <-fin:
	case <-time.After(time.Second):
	}
	mu.Lock()
	defer mu.Unlock()

	natural.FormatMessage(&msg)
	msg.Text = s.TransformReply(msg.Text)
	return msg
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

func (s *Service) Parse(ctx *natural.Context) (*natural.ParseResult, error) {
	var resLock sync.Mutex
	res := &natural.ParseResult{
		Text: ctx.Text,
	}

	r, _ := ParseSimple(ctx)
	if r != nil {
		res.Merge(r, 1)
	}

	r, _ = s.ParseRegular(ctx)
	if r != nil {
		res.Merge(r, 1)
	}

	wg := &sync.WaitGroup{}
	for name, p := range s.Cfg.Parsers {
		if p.Weight > 0 && (time.Now().Sub(p.LastSeen) < s.ParserKeepAlive) {
			wg.Add(1)
			go func(name string) {
				msg := sarif.CreateMessage("natural/parse", ctx)
				msg.Destination = name
				reply, ok := <-s.Request(msg)
				if !ok {
					return
				}
				r := &natural.ParseResult{}
				reply.DecodePayload(&r)
				resLock.Lock()
				res.Merge(r, 1)
				resLock.Unlock()
				wg.Done()
			}(name)
		}
	}

	waitTimeout(wg, time.Second)
	return res, nil
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

func ParseSimple(ctx *natural.Context) (*natural.ParseResult, error) {
	r := &natural.ParseResult{
		Text: ctx.Text,
	}

	if msg, ok := natural.ParseSimple(ctx.Text); ok {
		r.Intents = []*natural.Intent{{
			Type:    "simple",
			Intent:  msg.Action,
			Message: msg,
			Weight:  1,
		}}
	}
	return r, nil
}

func (s *Service) ParseRegular(ctx *natural.Context) (*natural.ParseResult, error) {
	r := &natural.ParseResult{
		Text: ctx.Text,
	}

	if msg, ok := s.regular.Parse(ctx.Text); ok {
		r.Intents = []*natural.Intent{{
			Type:    "regular",
			Intent:  msg.Action,
			Message: msg,
			Weight:  1,
		}}
		return r, nil
	}

	return r, nil
}

type msgLearn struct {
	Sentence string `json:"sentence"`
	Action   string `json:"action"`
}

func (s *Service) handleNaturalLearn(msg sarif.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	p.Sentence = strings.TrimSpace(p.Sentence)
	if p.Sentence == "" {
		return
	}
	p.Action = strings.TrimLeft(p.Action, "./")

	if err := s.regular.Learn(p.Sentence, p.Action); err != nil {
		s.ReplyBadRequest(msg, err)
	}
	s.Log("info", "associating '"+p.Sentence+"' with "+p.Action)
	s.Reply(msg, sarif.CreateMessage("natural/learned/meaning", p))
}
