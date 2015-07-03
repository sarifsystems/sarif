// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

type MeaningParser struct {
}

func NewMeaningParser() *MeaningParser {
	return &MeaningParser{}
}

func (p *MeaningParser) ParseImperative(tokens []*Token) (*Meaning, error) {
	m := &Meaning{
		Variables: make(map[string]string),
	}

	preps := make(map[string]int)

	it := newTokenIterator(tokens)
	t := it.Next()
	for t != nil {
		if t.Is("P") || t.Is("&") {
			preps[t.Lemma]++
		}
		t = it.Next()
	}

	it.Reset()
	t = it.Next()
	for t != nil {
		primary := true
		if m.Predicate == "" && t.Is("V") {
			m.Predicate = t.Lemma
			t = it.Next()
			continue
		}

		if primary {
			if t.Is("N") || t.Is("O") {
				m.Object = t.Lemma
				t = it.Next()
				primary = false
				continue
			}
			if t.Is("P") || t.Is("&") {
				primary = false
			}
		}

		if t.Is("P") || t.Is("&") {
			prep := t
			preps[t.Lemma]--
			values := make([]*Token, 0)
			t = it.Next()

			for t != nil {
				if t.Is("P") || t.Is("&") {
					preps[t.Lemma]--
					if preps[t.Lemma] <= 0 {
						break
					}
				}
				values = append(values, t)
				t = it.Next()
			}
			m.Variables[prep.Lemma] = JoinTokens(values)
			if len(values) > 1 {
				m.Variables[values[0].Lemma] = JoinTokens(values[1:])
			}
			continue
		}

		t = it.Next()
	}

	return m, nil
}

func (p *MeaningParser) ParseDeclarative(tokens []*Token) (*Meaning, error) {
	m := &Meaning{
		Variables: make(map[string]string),
	}

	it := newTokenIterator(tokens)
	t := it.Next()

	it.Reset()
	t = it.Next()

	predicate := false
	for t != nil {
		if m.Subject == "" && !predicate && (t.Is("O") || t.Is("N")) {
			m.Subject = t.Lemma
			t = it.Next()
			continue
		}

		if m.Predicate == "" && t.Is("V") {
			predicate = true
			m.Predicate = t.Lemma
			t = it.Next()
			continue
		}

		if m.Object == "" && t.Is("N") {
			m.Object = t.Lemma
			t = it.Next()
			continue
		}

		if t.Is("$") {
			m.Variables["value"] = t.Lemma
		}
		if t.Is("A") {
			m.Variables["adjective"] = t.Lemma
		}

		t = it.Next()
	}

	if m.Subject != "" {
		m.Variables["subject"] = m.Subject
	}
	if m.Predicate != "" {
		m.Variables["predicate"] = m.Predicate
		if m.Object != "" {
			m.Variables[m.Predicate] = m.Object
		}
	}
	if m.Object != "" {
		m.Variables["object"] = m.Object
	}

	return m, nil
}
