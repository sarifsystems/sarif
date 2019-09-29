// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlquery

import (
	"strings"
	"time"

	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/pkg/natural/query"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "nlquery",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ObjectLocations map[string]string
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	Config services.Config
	Cfg    Config
	sarif.Client

	parser       *query.Parser
	ObjectsTried map[string]bool
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		Config: deps.Config,
		Cfg: Config{
			ObjectLocations: make(map[string]string),
		},
		Client: deps.Client,

		parser:       query.NewParser(),
		ObjectsTried: make(map[string]bool),
	}
}

func (s *Service) Enable() error {
	s.Subscribe("natural/parse", "self", s.handleNaturalParse)
	s.Config.Get(&s.Cfg)
	s.parser.KnownObjects = s.Cfg.ObjectLocations
	return nil
}

func (s *Service) handleNaturalParse(msg sarif.Message) {
	ctx := &natural.Context{}
	msg.DecodePayload(ctx)
	if ctx.Text == "" {
		ctx.Text = msg.Text
	}

	res, err := s.parser.Parse(ctx)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	// Before we would try querying the generic store,
	// let's try to guess an objects location
	if len(res.Intents) > 0 && strings.HasPrefix(res.Intents[0].Message.Action, "store/scan/") {
		it := res.Intents[0]
		object := strings.TrimPrefix(it.Message.Action, "store/scan/")
		if ok := s.ObjectsTried[object]; !ok {
			s.ObjectsTried[object] = true
			singular := strings.Trim(object, "s")

			select {
			case <-s.Discover(object + "/list"):
				it.Message.Action = object + "/list"
				s.Cfg.ObjectLocations[object] = it.Message.Action
				s.Config.Set(&s.Cfg)
			case <-s.Discover(singular + "/list"):
				it.Message.Action = singular + "/list"
				s.Cfg.ObjectLocations[object] = it.Message.Action
				s.Config.Set(&s.Cfg)
			case <-time.After(time.Second):
			}
		}
	}

	s.Reply(msg, sarif.CreateMessage("natural/parsed", res))
}
