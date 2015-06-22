// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package reasoner

import (
	"strings"
	"time"
)

type Fact struct {
	Id int64 `json:"-"`

	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`

	SubjectType   string    `json:"subject_type" sql:"-"`
	PredicateType string    `json:"predicate_type" sql:"-"`
	ObjectType    string    `json:"object_type" sql:"-"`
	UpdatedAt     time.Time `json:"updated_at,omitempty" sql:"index"`
}

func (f Fact) String() string {
	return f.Subject + " " + f.Predicate + " " + f.Object + " ."
}

func (f *Fact) Each(fn func(string) string) {
	f.Subject = fn(f.Subject)
	f.Predicate = fn(f.Predicate)
	f.Object = fn(f.Object)
}

func GuessType(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") {
		return "uri"
	}
	if strings.Contains(s, " ") {
		return "literal"
	}
	if strings.Contains(s, ":") {
		return "uri"
	}
	return "literal"
}

func (f *Fact) FillMissingTypes() {
	if f.SubjectType == "" {
		f.SubjectType = GuessType(f.Subject)
	}
	if f.PredicateType == "" {
		f.PredicateType = GuessType(f.Predicate)
	}
	if f.ObjectType == "" {
		f.ObjectType = GuessType(f.Object)
	}
}

func GetLabelMappings(fs []*Fact) (map[string]string, []string) {
	mappings := make(map[string]string)
	for _, f := range fs {
		f.FillMissingTypes()
		if f.Predicate == "rdfs:label" {
			mappings[f.Subject] = f.Object
		}
		if f.SubjectType == "uri" {
			if _, ok := mappings[f.Subject]; !ok {
				mappings[f.Subject] = ""
			}
		}
		if f.PredicateType == "uri" {
			if _, ok := mappings[f.Predicate]; !ok {
				mappings[f.Predicate] = ""
			}
		}
		if f.ObjectType == "uri" {
			if _, ok := mappings[f.Object]; !ok {
				mappings[f.Object] = ""
			}
		}
	}
	missing := make([]string, 0)
	for m, v := range mappings {
		if v == "" {
			missing = append(missing, m)
		}
	}
	return mappings, missing
}

func getOrDefault(m map[string]string, s string) string {
	if v := m[s]; v != "" {
		return v
	}
	return s
}

func FormSentences(fs []*Fact) []string {
	m, _ := GetLabelMappings(fs)
	sentences := make([]string, 0)

	for _, f := range fs {
		if f.Predicate == "rdfs:label" {
			continue
		}
		s := getOrDefault(m, f.Subject)
		p := getOrDefault(m, f.Predicate)
		o := getOrDefault(m, f.Object)
		sentence := p + " of " + s + " is " + o + "."
		sentences = append(sentences, sentence)
	}
	return sentences
}

func ToJsonLd(fs []*Fact) []interface{} {
	subjects := make(map[string]map[string]interface{})
	nested := make(map[string]struct{})

	for _, f := range fs {
		if _, ok := subjects[f.Subject]; !ok {
			subjects[f.Subject] = map[string]interface{}{
				"@id": f.Subject,
			}
		}
	}

	for _, f := range fs {
		sub := subjects[f.Subject]
		pred := f.Predicate
		if pred == "rdf:type" {
			pred = "@type"
		}
		var obj interface{} = f.Object
		if sub, ok := subjects[f.Object]; ok {
			obj = sub
		}
		if existing, ok := sub[f.Predicate]; ok {
			switch v := existing.(type) {
			case []interface{}:
				sub[f.Predicate] = append(v, obj)
			default:
				sub[f.Predicate] = []interface{}{v, obj}
			}
		} else {
			sub[f.Predicate] = obj
		}
		nested[f.Object] = struct{}{}
	}

	for s := range nested {
		delete(subjects, s)
	}
	graph := make([]interface{}, 0, len(subjects))
	for _, s := range subjects {
		graph = append(graph, s)
	}
	return graph
}
