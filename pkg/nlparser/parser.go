// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlparser

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/xconstruct/stark/pkg/mlearning"
)

type Parser struct {
	Perceptron *mlearning.Perceptron
}

func New() *Parser {
	return &Parser{
		mlearning.NewPerceptron(),
	}
}

type Sentence struct {
	Words []string
	Tags  []string
}

const (
	START = "-START-"
	END   = "-END-"
)

func normalize(word string) string {
	return word
}

type sentenceSet struct {
	Sentences []Sentence

	CurrSentence, CurrWord int
	PrevClass, PrevClass2  string
	Context                []string
}

func (s *sentenceSet) Reset() {
	s.PrevClass, s.PrevClass2 = START, START
	s.CurrSentence, s.CurrWord = -1, 0
	for j := range s.Sentences {
		k := rand.Intn(j + 1)
		s.Sentences[j], s.Sentences[k] = s.Sentences[k], s.Sentences[j]
	}
}

func (s *sentenceSet) Next() bool {
	s.CurrWord++

	if s.CurrSentence == -1 || s.CurrWord >= len(s.Sentences[s.CurrSentence].Words) {
		s.CurrWord = 0
		s.CurrSentence++

		if s.CurrSentence >= len(s.Sentences) {
			return false
		}

		sentence := s.Sentences[s.CurrSentence]
		ctx := make([]string, len(sentence.Words)+2)
		ctx[0], ctx[len(ctx)-1] = START, END
		for i, w := range sentence.Words {
			ctx[i+1] = normalize(w)
		}
		s.Context = ctx
	}
	return true
}

func (s *sentenceSet) Class() mlearning.Class {
	return mlearning.Class(s.Sentences[s.CurrSentence].Tags[s.CurrWord])
}

func (s *sentenceSet) Features() []mlearning.Feature {
	i := s.CurrWord
	word := s.Sentences[s.CurrSentence].Words[s.CurrWord]
	wprev1, wprev2 := sget(s.Context, i-1), sget(s.Context, i-2)
	wnext1, wnext2 := sget(s.Context, i+1), sget(s.Context, i+2)
	fs := map[string]struct{}{}

	addFeat(fs, "bias")
	addFeat(fs, "i suffix", suffix(word))
	addFeat(fs, "i pref1", word[0:1])
	addFeat(fs, "i-1 tag", s.PrevClass)
	addFeat(fs, "i-2 tag", s.PrevClass2)
	addFeat(fs, "i tag+i-2 tag", s.PrevClass, s.PrevClass2)
	addFeat(fs, "i word", sget(s.Context, i))
	addFeat(fs, "i-1 tag+i word", s.PrevClass, word)
	addFeat(fs, "i-1 word", wprev1)
	addFeat(fs, "i-1 suffix", suffix(wprev1))
	addFeat(fs, "i-2 word", wprev2)
	addFeat(fs, "i+1 word", wnext1)
	addFeat(fs, "i+1 suffix", suffix(wnext1))
	addFeat(fs, "i+2 word", wnext2)

	fslice := make([]mlearning.Feature, 0, len(fs))
	for f := range fs {
		fslice = append(fslice, mlearning.Feature(f))
	}
	return fslice
}

func (s *sentenceSet) Predicted(c mlearning.Class) {
	s.PrevClass2, s.PrevClass = s.PrevClass, string(c)
}

func addFeat(fs map[string]struct{}, f string, args ...string) {
	if len(args) > 0 {
		f += "+" + strings.Join(args, "+")
	}
	fs[f] = struct{}{}
}

func suffix(s string) string {
	if l := len(s); l > 3 {
		return s[l-3:]
	}
	return s
}

func sget(sl []string, i int) string {
	if i > 0 && i < len(sl) {
		return sl[i]
	}
	return ""
}

func (p *Parser) Train(iterations int, sentences []Sentence) {
	set := &sentenceSet{
		Sentences: sentences,
	}
	for it := 0; it < iterations; it++ {
		set.Reset()
		c, n := p.Perceptron.Train(set)
		fmt.Printf("Iter %d: %d/%d=%.3f\n", it, c, n, float64(c)/float64(n)*100)
	}
}

func (p *Parser) Test(sentences []Sentence) {
	set := &sentenceSet{
		Sentences: sentences,
	}
	set.Reset()
	c, n := p.Perceptron.Test(set)
	fmt.Printf("Test: %d/%d=%.3f\n", c, n, float64(c)/float64(n)*100)
}
