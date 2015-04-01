// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mlearning

import "math/rand"

type Feature string
type Class string
type Weight float64

type Iterator interface {
	Next() bool
	Features() []Feature
	Class() Class
	Predicted(c Class)
}

type SimpleIterator struct {
	Index          int
	FeatureSlice   [][]Feature
	ClassSlice     []Class
	PredictedSlice []Class
}

func (i *SimpleIterator) Next() bool {
	i.Index++
	return i.Index <= len(i.FeatureSlice)
}

func (i *SimpleIterator) Features() []Feature {
	return i.FeatureSlice[i.Index-1]
}

func (i *SimpleIterator) Class() Class {
	return i.ClassSlice[i.Index-1]
}

func (i *SimpleIterator) Predicted(c Class) {
	if len(i.PredictedSlice) >= i.Index {
		i.PredictedSlice[i.Index-1] = c
	}
}

func (i *SimpleIterator) Reset() {
	i.Index = 0
	for j := range i.FeatureSlice {
		k := rand.Intn(j + 1)
		i.FeatureSlice[j], i.FeatureSlice[k] = i.FeatureSlice[k], i.FeatureSlice[j]
		i.ClassSlice[j], i.ClassSlice[k] = i.ClassSlice[k], i.ClassSlice[j]
	}
}

type Model struct {
	Weights map[Feature]map[Class]Weight
	Classes []Class

	I          int
	Totals     map[Feature]map[Class]Weight
	Timestamps map[Feature]map[Class]int
}

func NewModel() *Model {
	return &Model{
		Weights: make(map[Feature]map[Class]Weight),
		Classes: make([]Class, 0),

		I:          0,
		Totals:     make(map[Feature]map[Class]Weight),
		Timestamps: make(map[Feature]map[Class]int),
	}
}

type Perceptron struct {
	*Model
}

func NewPerceptron() *Perceptron {
	return &Perceptron{
		Model: NewModel(),
	}
}

func (p *Perceptron) Predict(features []Feature) (Class, Weight) {
	scores := map[Class]Weight{}
	for _, feat := range features {
		if weights, ok := p.Weights[feat]; ok {
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

func (p *Perceptron) Train(fs Iterator) (int, int) {
	c, n := 0, 0
	for fs.Next() {
		feats := fs.Features()
		guess, _ := p.Predict(feats)
		truth := fs.Class()
		fs.Predicted(guess)
		p.Update(truth, guess, feats)
		if guess == truth {
			c++
		}
		n++
	}
	return c, n
}

func (p *Perceptron) Test(fs Iterator) (int, int) {
	c, n := 0, 0
	for fs.Next() {
		feats := fs.Features()
		guess, _ := p.Predict(feats)
		fs.Predicted(guess)
		truth := fs.Class()
		if guess == truth {
			c++
		}
		n++
	}
	return c, n
}

func (p *Perceptron) Update(truth Class, guess Class, features []Feature) {
	p.I++
	for _, f := range features {
		p.UpdateFeature(truth, f, 1)
		p.UpdateFeature(guess, f, -1)
	}

}

func (p *Perceptron) UpdateFeature(c Class, f Feature, w Weight) {
	nrItersAtThisWeight := p.I
	if ci, ok := p.Timestamps[f]; ok {
		nrItersAtThisWeight -= ci[c]
	}

	if _, ok := p.Totals[f]; !ok {
		p.Totals[f] = map[Class]Weight{}
	}
	if _, ok := p.Weights[f]; !ok {
		p.Weights[f] = map[Class]Weight{}
	}
	if _, ok := p.Timestamps[f]; !ok {
		p.Timestamps[f] = map[Class]int{}
	}
	p.Totals[f][c] += Weight(nrItersAtThisWeight) * p.Weights[f][c]
	p.Weights[f][c] += w
	p.Timestamps[f][c] = p.I
}
