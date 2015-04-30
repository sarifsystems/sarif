// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sparql

import "testing"

func TestSimpleQuery(t *testing.T) {
	ep := NewEndpoint("http://dbpedia.org/sparql")

	var r ResourceResponse
	err := ep.Select(`
		select distinct *
		where {
			dbpedia:Tuomas_Holopainen dbpedia-owl:birthDate ?o
		} LIMIT 100
	`, &r)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestBuildQuery(t *testing.T) {
	var r ResourceResponse
	err := DBPedia.Query().
		Where("dbpedia:Tuomas_Holopainen", "dbpedia-owl:birthDate", "?o").
		Optional("dbpedia-owl:birthDate", "rdfs:label", "?rdfs_label").
		FilterLang("?rdfs_label", "EN").
		Exec(&r)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}

func TestBuildQueryDescribe(t *testing.T) {
	var r map[string]interface{}
	err := DBPedia.Query().
		Describe("dbpedia:Tuomas_Holopainen").
		Exec(&r)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)
}
