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
	words := make([]string, len(tokens))
	for i, t := range tokens {
		words[i] = t.Lemma
	}
	m := &Meaning{
		Variables: make(map[string]string),
		Words:     words,
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
				t.Tag("var")
				values = append(values, t)
				t = it.Next()
			}
			m.Variables[prep.Lemma] = JoinTokens(values)
			if len(values) > 1 && !values[0].Is("$") {
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
	it.Reset()
	t := it.Next()

	fact := true
	verb := false
	for t != nil {
		if !verb {
			// simple attributes: "[predicate] of [subject]"
			if m.Subject != "" && m.Predicate == "" && t.Is("P") {
				m.Predicate = m.Subject
				m.Subject = ""
				t = it.Next()
				continue
			}

			if couldBeNoun(t) {
				if m.Subject != "" {
					m.Subject += " "
				}
				m.Subject += t.Lemma
				t = it.Next()
				continue
			}

			if t.Is("V") {
				verb = true
				if t.Lemma != "is" && t.Lemma != "are" {
					fact = false
				}
				if m.Predicate == "" {
					m.Predicate = t.Lemma
				}
				t = it.Next()
				continue
			}
		}

		if t.Is("$") {
			m.Variables["value"] = t.Lemma
		}
		if t.Is("A") {
			m.Variables["adjective"] = t.Lemma
		}

		if couldBeNoun(t) {
			if m.Object != "" {
				m.Object += " "
			}
			m.Object += t.Lemma
			t = it.Next()
			continue
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
	if fact {
		m.Variables["fact"] = "true"
	}

	return m, nil
}

func (p *MeaningParser) ParseInterrogative(tokens []*Token) (*Meaning, error) {
	m := &Meaning{
		Variables: make(map[string]string),
	}

	it := newTokenIterator(tokens)
	it.Reset()
	t := it.Next()

	var q bool
	var query string
	for t != nil {
		// first interrogative pronoun
		if !q && t.Is("O") {
			q = true
			m.Variables["interrogative"] = t.Lemma
			t = it.Next()
			continue
		}

		// first verb is predicate
		if m.Predicate == "" && t.Is("V") {
			if t.Lemma != "is" && t.Lemma != "are" { // TODO: hard-coded
				m.Predicate = t.Lemma
			}
			// "what color does the car have?" -> color(car)
			if m.Subject != "" {
				m.Predicate = m.Subject
				m.Subject = ""
			}
			t = it.Next()
			continue
		}

		// rest of the sentence is query
		if query != "" {
			query += " "
		}
		query += t.Lemma

		// asking for simple attributes: "[predicate] of [subject]"
		if m.Subject != "" && m.Predicate == "" && t.Is("P") {
			m.Predicate = m.Subject
			m.Subject = ""
			t = it.Next()
			continue
		}

		if couldBeNoun(t) {
			if m.Subject != "" {
				m.Subject += " "
			}
			m.Subject += t.Lemma
			t = it.Next()
			continue
		}

		t = it.Next()
	}

	m.Variables["query"] = query
	if m.Subject != "" {
		m.Variables["subject"] = m.Subject
	}
	if m.Predicate != "" {
		m.Variables["predicate"] = m.Predicate
	}

	return m, nil
}

func couldBeNoun(t *Token) bool {
	return !t.Is("D") && !t.Is("P") && !t.Is("V")
}
