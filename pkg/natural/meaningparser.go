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
		Vars:   make([]*Var, 0),
		Tokens: tokens,
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

		if t.Is("P") || t.Is("&") || t.Is("T") {
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
			m.Vars = append(m.Vars, &Var{
				Name:  prep.Lemma,
				Value: JoinTokens(values),
			})
			if len(values) > 1 && !values[0].Is("$") {
				m.Vars = append(m.Vars, &Var{
					Name:  values[0].Lemma,
					Value: JoinTokens(values[1:]),
				})
			}
			continue
		}

		t = it.Next()
	}

	return m, nil
}

func (p *MeaningParser) ParseDeclarative(tokens []*Token) (*Meaning, error) {
	m := &Meaning{
		Vars: make([]*Var, 0),
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
			m.Vars = append(m.Vars, &Var{
				Name:  "value",
				Value: t.Lemma,
			})
		}
		if t.Is("A") {
			m.Vars = append(m.Vars, &Var{
				Name:  "adjective",
				Value: t.Lemma,
			})
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
		m.Vars = append(m.Vars, &Var{
			Name:  "subject",
			Value: m.Subject,
		})
	}
	if m.Predicate != "" {
		m.Vars = append(m.Vars, &Var{
			Name:  "predicate",
			Value: m.Predicate,
		})
		if m.Object != "" {
			m.Vars = append(m.Vars, &Var{
				Name:  m.Predicate,
				Value: m.Object,
			})
		}
	}
	if m.Object != "" {
		m.Vars = append(m.Vars, &Var{
			Name:  "object",
			Value: m.Object,
		})
	}
	if fact {
		m.Vars = append(m.Vars, &Var{
			Name:  "fact",
			Value: "true",
		})
	}

	return m, nil
}

func (p *MeaningParser) ParseInterrogative(tokens []*Token) (*Meaning, error) {
	m := &Meaning{
		Vars: make([]*Var, 0),
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
			m.Vars = append(m.Vars, &Var{
				Name:  "interrogative",
				Value: t.Lemma,
			})
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

	m.Vars = append(m.Vars, &Var{
		Name:  "query",
		Value: query,
	})
	m.Vars = append(m.Vars, &Var{
		Name:  "subject",
		Value: m.Subject,
	})
	m.Vars = append(m.Vars, &Var{
		Name:  "predicate",
		Value: m.Predicate,
	})

	return m, nil
}

func couldBeNoun(t *Token) bool {
	return !t.Is("D") && !t.Is("P") && !t.Is("V")
}
