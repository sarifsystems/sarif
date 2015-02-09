// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlparser

import (
	"fmt"
	"math/rand"
	"strings"
)

type Feature string
type Class string
type Weight float64

type Model struct {
	Weights map[Feature]map[Class]Weight
}

type Parser struct {
	weights map[Feature]map[Class]Weight
	classes []Class

	i          int
	totals     map[Feature]map[Class]Weight
	timestamps map[Feature]map[Class]int
}

func New() *Parser {
	return &Parser{
		make(map[Feature]map[Class]Weight),
		make([]Class, 0),

		0,
		make(map[Feature]map[Class]Weight),
		make(map[Feature]map[Class]int),
	}
}

func (p *Parser) Predict(features []Feature) (Class, Weight) {
	scores := map[Class]Weight{}
	for _, feat := range features {
		if weights, ok := p.weights[feat]; ok {
			for class, weight := range weights {
				scores[class] += weight
			}
		}
	}

	mclass := Class("")
	mweight := Weight(0)
	for class, weight := range scores {
		if weight > mweight {
			mclass, mweight = class, weight
		}
	}
	return mclass, mweight
}

type Sentence struct {
	Words []string
	Tags  []Class
}

const (
	START = "-START-"
	END   = "-END-"
)

func normalize(word string) string {
	return word
}

func (p *Parser) Train(iterations int, sentences []Sentence) {
	var prev, prev2 Class = START, START

	for it := 0; it < iterations; it++ {
		c, n := 0, 0
		for _, s := range sentences {
			context := make([]string, len(s.Words)+2)
			context[0], context[len(context)-1] = START, END
			for i, w := range s.Words {
				context[i] = normalize(w)
			}

			for i, word := range s.Words {
				feats := p.GetFeatures(i, word, context, prev, prev2)
				guess, _ := p.Predict(feats)
				p.update(s.Tags[i], guess, feats)
				prev2, prev = prev, guess
				if guess == s.Tags[i] {
					c++
				}
				n++
			}
		}

		for j := range sentences {
			k := rand.Intn(j + 1)
			sentences[j], sentences[k] = sentences[k], sentences[j]
		}
		fmt.Printf("Iter %d: %d/%d=%.3f\n", it, c, n, float64(c)/float64(n)*100)
	}
}

func (p *Parser) Test(sentences []Sentence) {
	var prev, prev2 Class = START, START

	c, n := 0, 0
	for _, s := range sentences {
		context := make([]string, len(s.Words)+2)
		context[0], context[len(context)-1] = START, END
		for i, w := range s.Words {
			context[i] = normalize(w)
		}

		for i, word := range s.Words {
			feats := p.GetFeatures(i, word, context, prev, prev2)
			guess, _ := p.Predict(feats)
			prev2, prev = prev, guess
			if guess == s.Tags[i] {
				c++
			}
			n++
		}
	}

	for j := range sentences {
		k := rand.Intn(j + 1)
		sentences[j], sentences[k] = sentences[k], sentences[j]
	}
	fmt.Printf("%d/%d=%.3f\n", c, n, float64(c)/float64(n)*100)
}

func (p *Parser) update(truth Class, guess Class, features []Feature) {
	p.i++
	for _, f := range features {
		p.updateFeature(truth, f, 1)
		p.updateFeature(guess, f, -1)
	}

}

func (p *Parser) updateFeature(c Class, f Feature, w Weight) {
	nrItersAtThisWeight := p.i
	if ci, ok := p.timestamps[f]; ok {
		nrItersAtThisWeight -= ci[c]
	}

	if _, ok := p.totals[f]; !ok {
		p.totals[f] = map[Class]Weight{}
	}
	if _, ok := p.weights[f]; !ok {
		p.weights[f] = map[Class]Weight{}
	}
	if _, ok := p.timestamps[f]; !ok {
		p.timestamps[f] = map[Class]int{}
	}
	p.totals[f][c] += Weight(nrItersAtThisWeight) * p.weights[f][c]
	p.weights[f][c] += w
	p.timestamps[f][c] = p.i
}

func addFeat(fs map[Feature]struct{}, f string, args ...string) {
	if len(args) > 0 {
		f += "+" + strings.Join(args, "+")
	}
	fs[Feature(f)] = struct{}{}
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

func (p *Parser) GetFeatures(i int, word string, context []string, prev, prev2 Class) []Feature {
	wprev1, wprev2 := sget(context, i-1), sget(context, i-2)
	wnext1, wnext2 := sget(context, i+1), sget(context, i+2)
	fs := map[Feature]struct{}{}

	addFeat(fs, "bias")
	addFeat(fs, "i suffix", suffix(word))
	addFeat(fs, "i pref1", word[0:1])
	addFeat(fs, "i-1 tag", string(prev))
	addFeat(fs, "i-2 tag", string(prev2))
	addFeat(fs, "i tag+i-2 tag", string(prev), string(prev2))
	addFeat(fs, "i word", sget(context, i))
	addFeat(fs, "i-1 tag+i word", string(prev), word)
	addFeat(fs, "i-1 word", wprev1)
	addFeat(fs, "i-1 suffix", suffix(wprev1))
	addFeat(fs, "i-2 word", wprev2)
	addFeat(fs, "i+1 word", wnext1)
	addFeat(fs, "i+1 suffix", suffix(wnext1))
	addFeat(fs, "i+2 word", wnext2)

	fslice := make([]Feature, 0, len(fs))
	for f := range fs {
		fslice = append(fslice, f)
	}
	return fslice
}

func (p *Parser) GetModel() *Model {
	return &Model{p.weights}
}

func (p *Parser) SetModel(m *Model) {
	p.weights = m.Weights
}
