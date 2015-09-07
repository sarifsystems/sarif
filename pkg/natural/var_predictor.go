// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xconstruct/stark/pkg/mlearning"
)

type VarPredictor struct {
	Tokenizer  *Tokenizer
	Perceptron *mlearning.Perceptron
}

func NewVarPredictor(tok *Tokenizer) *VarPredictor {
	return &VarPredictor{
		tok,
		mlearning.NewPerceptron(),
	}
}

type varFeature struct {
	Sentence []*Token
	Action   string
	Pos      int
}

func (f *varFeature) Features() []mlearning.Feature {
	fs := map[string]struct{}{}

	addFeat(fs, "bias")
	if f.Pos > 0 {
		addFeat(fs, "1 word", f.Sentence[0].Lemma)
		addFeat(fs, "i-1 word", f.Sentence[f.Pos-1].Lemma)
	}
	if f.Pos > 1 {
		addFeat(fs, "2 word", f.Sentence[1].Lemma)
		addFeat(fs, "i-2 word", f.Sentence[f.Pos-2].Lemma)
	}
	if f.Action != "" {
		addFeat(fs, "action", f.Action)

		parts := strings.SplitN(f.Action, "/", 4)
		for i := 0; i < len(parts); i++ {
			addFeat(fs, strconv.Itoa(i)+" action", strings.Join(parts[0:i], "/"))
		}
	}

	fslice := make([]mlearning.Feature, 0, len(fs))
	for f := range fs {
		fslice = append(fslice, mlearning.Feature(f))
	}
	return fslice
}

func (p *VarPredictor) dataToIterator(dataset DataSet) *mlearning.SimpleIterator {
	vs := &mlearning.SimpleIterator{}

	for _, data := range dataset {
		sen := data.CleanedSentence("[name]")
		tok := p.Tokenizer.Tokenize(sen)

		for i, t := range tok {
			if !strings.HasPrefix(t.Value, "[") {
				continue
			}

			name := strings.Trim(t.Value, "[]")
			vs.FeatureSlice = append(vs.FeatureSlice, &varFeature{
				Sentence: tok,
				Action:   data.Action,
				Pos:      i,
			})
			vs.ClassSlice = append(vs.ClassSlice, mlearning.Class(name))
		}
	}

	return vs
}

func (p *VarPredictor) Train(iterations int, dataset DataSet, tok *Tokenizer) {
	set := p.dataToIterator(dataset)
	for it := 0; it < iterations; it++ {
		set.Reset(true)
		c, n := p.Perceptron.Train(set)
		fmt.Printf("VarPredictor iter %d: %d/%d=%.3f\n", it, c, n, float64(c)/float64(n)*100)
	}
}

func (p *VarPredictor) Test(dataset DataSet) {
	set := p.dataToIterator(dataset)
	set.Reset(true)
	c, n := p.Perceptron.Test(set)
	fmt.Printf("Test: %d/%d=%.3f\n", c, n, float64(c)/float64(n)*100)
}

func (p *VarPredictor) Predict(s string, action string, pos int) (string, float64) {
	tok := p.Tokenizer.Tokenize(s)

	set := &mlearning.SimpleIterator{}
	set.FeatureSlice = append(set.FeatureSlice, &varFeature{
		Sentence: tok,
		Action:   action,
		Pos:      pos,
	})
	set.Reset(false)
	set.Next()
	guess, w := p.Perceptron.Predict(set.Features())
	return string(guess), float64(w)
}

func (p *VarPredictor) PredictTokens(tok []*Token, action string) []*Var {
	vs := make([]*Var, 0)

	prevVar := false
	set := &mlearning.SimpleIterator{}
	for i, t := range tok {
		if !t.Is("var") || prevVar {
			prevVar = false
			continue
		}

		prevVar = true
		set.FeatureSlice = append(set.FeatureSlice, &varFeature{
			Sentence: tok,
			Action:   action,
			Pos:      i,
		})
	}

	set.Reset(false)
	for set.Next() {
		name, w := p.Perceptron.Predict(set.Features())
		pos := set.FeatureSlice[set.Index-1].(*varFeature).Pos
		i := pos
		for i < len(tok) && tok[i].Is("var") {
			i++
		}

		vs = append(vs, &Var{
			Name:   string(name),
			Value:  JoinTokens(tok[pos:i]),
			Weight: float64(w),
		})
	}

	return vs
}
