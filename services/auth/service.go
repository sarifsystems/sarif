// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service auth provides token and challenge based auth.
package auth

import (
	"crypto/rand"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "auth",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	Config Config
	cfg    services.Config
	sarif.Client

	SessionTokens map[string]time.Time
}

type Config struct {
	Tokens map[string]bool
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Client:        deps.Client,
		cfg:           deps.Config,
		SessionTokens: make(map[string]time.Time),
	}
	return s
}

func (s *Service) Enable() error {
	s.Config.Tokens = make(map[string]bool)
	s.cfg.Get(&s.Config)

	s.Subscribe("proto/hi", "", s.handleAuthRequest)
	s.Subscribe("auth/new/token", "", s.handleAuthToken)
	s.Subscribe("auth/new/otp", "", s.handleAuthOtp)
	return nil
}

func (s *Service) handleAuthRequest(msg sarif.Message) {
	var ci sarif.ClientInfo
	msg.DecodePayload(&ci)
	if ci.Auth == "" {
		return
	}

	authed := false
	auths := strings.Split(ci.Auth, ",")
	for _, auth := range auths {
		if s.Config.Tokens[auth] || s.SessionTokens[auth].After(time.Now()) {
			authed = true
		}
	}
	if !authed {
		return
	}

	s.Reply(msg, sarif.CreateMessage("proto/allow", nil))
}

func (s *Service) handleAuthToken(msg sarif.Message) {
	tok := "token/std:" + sarif.GenerateId() + sarif.GenerateId() + sarif.GenerateId()
	s.Config.Tokens[tok] = true
	s.cfg.Set(s.Config)

	s.Reply(msg, sarif.CreateMessage("auth/generated", sarif.ClientInfo{
		Auth: tok,
	}))
}

func (s *Service) handleAuthOtp(msg sarif.Message) {
	tok := "otp/std:" + GenerateDigits()
	s.SessionTokens[tok] = time.Now().Add(time.Minute)

	s.Reply(msg, sarif.CreateMessage("auth/generated", sarif.ClientInfo{
		Auth: tok,
	}))
}

func GenerateDigits() string {
	const num = "123456789"
	var bytes = make([]byte, 6)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = num[b%byte(len(num))]
	}
	return string(bytes)
}
