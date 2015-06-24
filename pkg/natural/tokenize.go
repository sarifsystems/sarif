// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"bufio"
	"io"
	"strings"
)

type Sentence []*Token

type Token struct {
	Value string              `json:"value,omitempty"`
	Lemma string              `json:"lemma,omitempty"`
	Tags  map[string]struct{} `json:"tags,omitempty"`
}

func (t Token) Is(tag string) bool {
	_, ok := t.Tags[tag]
	return ok
}

type Tokenizer struct {
	SplitQuoted bool
	StopWords   []string
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		SplitQuoted: true,
		StopWords:   DefaultStopWords,
	}
}

func (t *Tokenizer) Tokenize(s string) []*Token {
	s = strings.ToLower(s)
	s = strings.TrimRight(s, ".!? ")

	var words []string
	if t.SplitQuoted {
		words, _ = SplitQuoted(s, " ")
	} else {
		words = strings.Split(s, " ")
	}

	tokens := make([]*Token, 0, len(words))
	for _, w := range words {
		tok := &Token{
			Value: w,
			Lemma: w,
			Tags:  make(map[string]struct{}),
		}
		if inStringSlice(w, t.StopWords) {
			tok.Tags["STOPWORD"] = struct{}{}
		}

		tokens = append(tokens, tok)
	}

	return tokens
}

func LoadCoNLL(r io.Reader) ([]Sentence, error) {
	sentences := make([]Sentence, 0)
	scan := bufio.NewScanner(r)
	tokens := make([]*Token, 0)
	for scan.Scan() {
		if scan.Text() == "" {
			sentences = append(sentences, tokens)
			tokens = make([]*Token, 0)
		} else {
			parts := strings.Split(scan.Text(), "\t")
			t := &Token{
				Value: parts[0],
				Lemma: parts[0],
				Tags:  make(map[string]struct{}),
			}
			t.Tags[parts[1]] = struct{}{}
			tokens = append(tokens, t)
		}
	}
	return sentences, scan.Err()
}
