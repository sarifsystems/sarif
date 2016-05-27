// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service reasoner provides a knowledge base and inference engine.
package reasoner

import (
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/sarifsystems/sarif/pkg/sparql"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "reasoner",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *gorm.DB
	Log    sarif.Logger
	Client *sarif.Client
}

type Service struct {
	DB  *gorm.DB
	Log sarif.Logger
	*sarif.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() error {
	createIndizes := !s.DB.HasTable(&Fact{})
	if err := s.DB.AutoMigrate(&Fact{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Fact{}).AddIndex("subject_predicate", "subject", "predicate").Error; err != nil {
			return err
		}
		if err := s.DB.Model(&Fact{}).AddIndex("predicate_object", "predicate", "object").Error; err != nil {
			return err
		}
	}

	s.Subscribe("concepts/query", "", s.HandleQuery)
	s.Subscribe("concepts/query_external", "", s.HandleQueryExternal)
	s.Subscribe("concepts/store", "", s.HandleStore)
	s.Subscribe("concept", "", s.HandleStore)
	return nil
}

type resultPayload struct {
	Result interface{} `json:"result"`
	Facts  []*Fact     `json:"facts"`
}

func (p resultPayload) Text() string {
	if p.Facts == nil || len(p.Facts) == 0 {
		return "No facts."
	}
	s := strings.Join(FormSentences(p.Facts), "\n")
	if s != "" {
		return s
	}
	for _, f := range p.Facts {
		if s != "" {
			s += "\n"
		}
		s += f.String()
	}
	return s
}

func (s *Service) HandleQuery(msg sarif.Message) {
	var f Fact
	if err := msg.DecodePayload(&f); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	f, err := s.InterpretLiterals(f)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	var facts []*Fact
	if err := s.DB.Where(f).Limit(100).Find(&facts).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	if facts, err = s.AddLabelFacts(facts); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("concepts/result", &resultPayload{
		ToJsonLd(facts),
		facts,
	}))
}

func (s *Service) InterpretLiterals(f Fact) (Fact, error) {
	f.FillMissingTypes()
	literals := make([]string, 0)
	if f.SubjectType == "literal" {
		literals = append(literals, f.Subject)
	}
	if f.PredicateType == "literal" {
		literals = append(literals, f.Predicate)
	}
	if f.ObjectType == "literal" {
		literals = append(literals, f.Object)
	}
	if len(literals) == 0 {
		return f, nil
	}

	var results []*Fact
	err := s.DB.
		Where("predicate = ?", "rdfs:label").
		Where("object IN (?)", literals).
		Find(&results).
		Error
	if err != nil {
		return f, err
	}
	for _, r := range results {
		f.EachType(func(s, t string) (string, string) {
			if s == r.Object {
				return r.Subject, "uri"
			}
			return s, t
		})
	}
	return f, nil
}

func (s *Service) HandleStore(msg sarif.Message) {
	var f Fact
	if err := msg.DecodePayload(&f); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	f, err := s.InterpretLiterals(f)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	db := s.DB.Where(Fact{Subject: f.Subject, Predicate: f.Predicate})
	db = s.DB.Assign(f)
	if err := db.FirstOrCreate(&f).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, sarif.CreateMessage("concepts/stored", &f))
}

func (s *Service) HandleQueryExternal(msg sarif.Message) {
	var f Fact
	if err := msg.DecodePayload(&f); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	facts := []*Fact{&f}
	FillVariables(facts)

	var r sparql.ResourceResponse
	q := sparql.DBPedia.Query()
	q = BuildQuery(q, facts)
	if err := q.Exec(&r); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	result := ApplyBindings(facts, r.Results.Bindings, sparql.CommonPrefixes)
	s.Reply(msg, sarif.CreateMessage("concepts/result", &resultPayload{
		ToJsonLd(result),
		result,
	}))

	for _, f := range result {
		if err := s.DB.FirstOrCreate(&f, &f).Error; err != nil {
			s.Log.Errorln("[reasoner] error updating external fact:", err)
		}
	}
}

func (s *Service) AddLabelFacts(fs []*Fact) ([]*Fact, error) {
	_, missing := GetLabelMappings(fs)
	if len(missing) == 0 {
		return fs, nil
	}

	var results []*Fact
	err := s.DB.
		Where("predicate = ?", "rdfs:label").
		Where("subject IN (?)", missing).
		Find(&results).
		Error
	if err != nil {
		return fs, err
	}
	return append(fs, results...), nil
}

func FillVariables(fs []*Fact) {
	for i, f := range fs {
		if f.Subject == "" {
			f.Subject = "?s_gen" + strconv.Itoa(i)
		}
		if f.Predicate == "" {
			f.Predicate = "?p_gen" + strconv.Itoa(i)
		}
		if f.Object == "" {
			f.Object = "?o_gen" + strconv.Itoa(i)
		}
	}
}

func BuildQuery(q *sparql.Query, facts []*Fact) *sparql.Query {
	for _, f := range facts {
		q = q.Where(f.Subject, f.Predicate, f.Object)
		if f.Predicate == "rdfs:label" {
			q = q.FilterLang(f.Object, "EN")
		}
	}
	return q
}

func ApplyBindings(fs []*Fact, bs []sparql.Row, ns map[string]string) []*Fact {
	result := make([]*Fact, 0, len(fs))

	for i, br := range bs {
		for _, fOrig := range fs {
			f := *fOrig
			modified := false
			for v, b := range br {
				v = "?" + v
				value := b.Value
				if b.Type == "uri" {
					value = AddNamespacePrefix(value, ns)
				}
				if f.Subject == v {
					f.Subject = value
					f.SubjectType = b.Type
					modified = true
				}
				if f.Predicate == v {
					f.Predicate = value
					f.PredicateType = b.Type
					modified = true
				}
				if f.Object == v {
					f.Object = value
					f.ObjectType = b.Type
					modified = true
				}
			}
			if i == 0 || modified {
				result = append(result, &f)
			}
		}
	}

	return result
}

func AddNamespacePrefix(s string, ns map[string]string) string {
	for prefix, uri := range ns {
		if strings.HasPrefix(s, uri) {
			return prefix + ":" + strings.TrimPrefix(s, uri)
		}
	}
	return s
}

func CreateURIFromLiteral(s string) string {
	return "sarif:" + strings.Replace(s, " ", "_", -1)
}
