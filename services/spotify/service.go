package spotify

import (
	"errors"
	"time"

	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const state = "sarif123"

var Module = &services.Module{
	Name:        "spotify",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ClientSecrets
	Token *oauth2.Token
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	Config Config
	cfg    services.Config
	sarif.Client
	Spotify *spotify.Client

	authInProgress *ClientSecrets
	CurrentState   spotify.PlayerState
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Config: Config{},
		Client: deps.Client,
		cfg:    deps.Config,
	}
	return s
}

func (s *Service) Enable() error {
	s.cfg.Get(&s.Config)

	s.Subscribe("spotify/authenticate", "", s.handleAuthenticate)
	s.Subscribe("spotify/auth/confirm", "", s.handleAuthConfirm)

	if s.Config.Token != nil {
		s.init()
	}

	return nil
}

type ClientSecrets struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

func (s *Service) handleAuthenticate(msg sarif.Message) {
	var secrets ClientSecrets
	msg.DecodePayload(&secrets)
	if secrets.ClientID == "" || secrets.ClientSecret == "" {
		s.ReplyBadRequest(msg, errors.New("Missing client id or secret"))
		return
	}

	if secrets.RedirectURI == "" {
		s.ReplyBadRequest(msg, errors.New("Missing redirect uri - example: http://localhost/api/v0/spotify/auth/confirm?authtoken=xxxxx"))
		return
	}

	s.authInProgress = &secrets

	url := secrets.authenticator().AuthURL(state)
	reply := sarif.CreateMessage("spotify/handshake", map[string]string{
		"url": url,
	})
	reply.Text = "Please visit: " + url
	s.Reply(msg, reply)
}

type AuthConfirmPayload struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func stateToInfo(state spotify.PlayerState) schema.MusicInfo {
	info := schema.MusicInfo{}
	info.IsPlaying = state.Playing
	if state.Playing {
		if len(state.Item.Artists) > 0 {
			info.Artist = state.Item.Artists[0].Name
		}
		info.Device = state.Device.Name
		info.Track = state.Item.Name
		info.Album = state.Item.Album.Name
		info.Duration = state.Item.Duration / 1000

		t := int64(state.Timestamp) - int64(state.Progress)
		info.Time = time.Unix(t/1000, 0).Local()
	}
	return info
}

func (s *Service) handleAuthConfirm(msg sarif.Message) {
	var p AuthConfirmPayload
	msg.DecodePayload(&p)

	if s.authInProgress == nil {
		s.ReplyInternalError(msg, errors.New("No auth in progress"))
		return
	}
	secrets := *s.authInProgress

	tok, err := secrets.authenticator().Exchange(p.Code)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Config.ClientSecrets = secrets
	s.Config.Token = tok
	s.authInProgress = nil
	s.cfg.Set(s.Config)

	s.init()
}

func (s *Service) init() {
	isRunning := s.Spotify != nil
	client := s.Config.authenticator().NewClient(s.Config.Token)
	client.AutoRetry = true
	s.Spotify = &client

	if !isRunning {
		go s.readLoop()
	}
}

func (s *Service) readLoop() {
	country := "DE"
	opt := &spotify.Options{
		Country: &country,
	}

	for {
		state, err := s.Spotify.PlayerStateOpt(opt)
		if err != nil {
			s.Log("err/internal", err.Error())
			time.Sleep(5 * time.Minute)
			continue
		}

		dur := s.AdvanceState(*state)
		time.Sleep(dur)
	}
}

func (s *Service) AdvanceState(state spotify.PlayerState) time.Duration {
	prev := s.CurrentState
	s.CurrentState = state

	next := 1 * time.Minute

	if state.Playing {
		if !prev.Playing {
			s.PublishState("started", state)
		} else {
			// If track changed or user rewinds
			if state.Item.ID != prev.Item.ID || state.Progress < prev.Progress {
				s.maybeScrobble(prev)
				s.PublishState("changed", state)
			}
		}

		// If near the end of a song, try to find a good time for scrobbling
		remaining := time.Duration(state.Item.Duration-state.Progress) * time.Millisecond
		remaining -= 30 * time.Second
		if remaining < next && remaining > 5*time.Second {
			next = remaining
		}
	} else {
		if prev.Playing {
			s.maybeScrobble(prev)
			s.PublishState("stopped", state)
		}
		next = 3 * time.Minute
	}

	return next
}

func (s *Service) PublishState(action string, state spotify.PlayerState) {
	info := stateToInfo(state)
	s.Publish(sarif.CreateMessage("spotify/playback/"+action, state))
	s.Publish(sarif.CreateMessage("music/"+action, info))
}

func (s *Service) maybeScrobble(state spotify.PlayerState) {
	if state.Item.Duration < 30*1e3 {
		return
	}

	pct := float32(state.Progress) / float32(state.Item.Duration)
	if pct >= 0.5 || state.Progress >= 4*60*1000 {
		s.PublishState("scrobble", state)
	}
}

func (s ClientSecrets) authenticator() spotify.Authenticator {
	auth := spotify.NewAuthenticator(s.RedirectURI, spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	auth.SetAuthInfo(s.ClientID, s.ClientSecret)
	return auth
}
