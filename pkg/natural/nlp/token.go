// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

type Token struct {
	Value string              `json:"value,omitempty"`
	Lemma string              `json:"lemma,omitempty"`
	Tags  map[string]struct{} `json:"tags,omitempty"`
}

func (t Token) Is(tag string) bool {
	_, ok := t.Tags[tag]
	return ok
}

func (t *Token) Tag(tag string) {
	t.Tags[tag] = struct{}{}
}

type tokenIterator struct {
	tokens []*Token
	pos    int
}

func newTokenIterator(tokens []*Token) *tokenIterator {
	return &tokenIterator{tokens, -1}
}

func (it *tokenIterator) Peek() *Token {
	if it.pos+1 >= len(it.tokens) {
		return nil
	}
	return it.tokens[it.pos+1]
}

func (it *tokenIterator) Next() *Token {
	it.pos++
	if it.pos >= len(it.tokens) {
		return nil
	}
	return it.tokens[it.pos]
}

func (it *tokenIterator) Reset() {
	it.pos = -1
}

func JoinTokens(ts []*Token) string {
	s := ""
	for i, t := range ts {
		if i > 0 {
			s += " "
		}
		s += t.Value
	}
	return s
}
