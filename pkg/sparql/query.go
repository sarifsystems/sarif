// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sparql

import (
	"errors"
	"strings"
)

type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

type Endpoint interface {
	Select(q string, result interface{}) error
	Describe(q string, result interface{}) error
}

func (t Triple) String() string {
	return t.Subject + " " + t.Predicate + " " + t.Object
}

type Query struct {
	endpoint Endpoint

	qtype    string
	prefixes map[string]string
	fields   []string
	where    []string
}

func New(ep ...Endpoint) *Query {
	q := &Query{
		qtype:    "select",
		prefixes: make(map[string]string),
		fields:   make([]string, 0),
		where:    make([]string, 0),
	}
	if len(ep) > 0 {
		q.endpoint = ep[0]
	}
	return q
}

func (q *Query) Prefix(p, uri string) *Query {
	q.prefixes[p] = uri
	return q
}

func (q *Query) Where(s, p, o string) *Query {
	t := Triple{s, p, o}
	q.where = append(q.where, t.String())
	return q
}

func (q *Query) Filter(fs ...string) *Query {
	for _, f := range fs {
		q.where = append(q.where, " FILTER "+f)
	}
	return q
}

func (q *Query) FilterLang(v, lang string) *Query {
	return q.Filter("langMatches(lang(" + v + `), "` + lang + `")`)
}

func (q *Query) Optional(s, p, o string) *Query {
	t := Triple{s, p, o}
	q.where = append(q.where, "OPTIONAL { "+t.String()+" }")
	return q
}

func (q *Query) String() string {
	s := ""
	for p, uri := range q.prefixes {
		s += " PREFIX " + p + ": <" + uri + ">"
	}

	switch q.qtype {
	case "describe":
		s += " DESCRIBE"
	default:
		s += " SELECT"
	}
	if len(q.fields) > 0 {
		s += " " + strings.Join(q.fields, " ")
	} else {
		s += " *"
	}

	if len(q.where) > 0 {
		s += " WHERE {"
		for _, w := range q.where {
			s += " " + w + " ."
		}
		s += " }"
	}

	return s
}

func (q *Query) Select(fields ...string) *Query {
	q.qtype = "select"
	q.fields = append(q.fields, fields...)
	return q
}

func (q *Query) Describe(fields ...string) *Query {
	q.qtype = "describe"
	q.fields = append(q.fields, fields...)
	return q
}

func (q *Query) Exec(result interface{}) error {
	if q.endpoint == nil {
		return errors.New("No SPARQL endpoint given")
	}
	switch q.qtype {
	case "select":
		return q.endpoint.Select(q.String(), result)
	case "describe":
		return q.endpoint.Describe(q.String(), result)
	}
	panic("Unknown query type: " + q.qtype)
}

var (
	CommonPrefixes = map[string]string{
		"sarif": "sarif://schema/",

		"dbpedia":      "http://dbpedia.org/resource/",
		"dbpedia-owl":  "http://dbpedia.org/ontology/",
		"dbpedia-prop": "http://dbpedia.org/property/",
		"foaf":         "http://xmlns.com/foaf/0.1/",
		"freebase-key": "http://rdf.freebase.com/key/",
		"freebase-ns":  "http://rdf.freebase.com/ns/",
		"mo":           "http://purl.org/ontology/mo/",
		"owl":          "http://www.w3.org/2002/07/owl#",
		"rdf":          "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"rdfs":         "http://www.w3.org/2000/01/rdf-schema#",
		"schema":       "http://schema.org/",
		"vcard":        "http://www.w3.org/2006/vcard/ns#",
		"yago":         "http://yago-knowledge.org/resource/",
		"xsd":          "http://www.w3.org/2001/XMLSchema#",
	}
)
