// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/xconstruct/stark/pkg/mlearning"
	"github.com/xconstruct/stark/pkg/nlparser"
)

var (
	modelPath = flag.String("model", "model.gob", "model file to save/load")
)

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "train":
		train()
	case "test":
		test()
	case "tag":
		tag()
	default:
		log.Fatal("train/tag")
	}
}

func train() {
	fpath := flag.Arg(1)
	f, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}

	sentences := make([]nlparser.Sentence, 0)
	train := bufio.NewScanner(f)
	words := make([]string, 0)
	tags := make([]string, 0)
	for train.Scan() {
		if train.Text() == "" {
			sentences = append(sentences, nlparser.Sentence{words, tags})
			words = make([]string, 0)
			tags = make([]string, 0)
		} else {
			parts := strings.Split(train.Text(), "\t")
			words = append(words, parts[0])
			tags = append(tags, parts[1])
		}
	}
	f.Close()
	if err := train.Err(); err != nil {
		log.Fatal(err)
	}

	p := nlparser.New()
	p.Train(5, sentences)

	fm, err := os.Create(*modelPath)
	if err != nil {
		log.Fatal(err)
	}
	enc := gob.NewEncoder(fm)
	if err := enc.Encode(p.Perceptron.Model); err != nil {
		log.Fatal(err)
	}
}

func test() {
	p := nlparser.New()

	fm, err := os.Open(*modelPath)
	if err != nil {
		log.Fatal(err)
	}
	dec := gob.NewDecoder(fm)
	model := &mlearning.Model{}
	if err := dec.Decode(model); err != nil {
		log.Fatal(err)
	}
	p.Perceptron.Model = model

	fpath := flag.Arg(1)
	f, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}

	sentences := make([]nlparser.Sentence, 0)
	train := bufio.NewScanner(f)
	words := make([]string, 0)
	tags := make([]string, 0)
	for train.Scan() {
		if train.Text() == "" {
			sentences = append(sentences, nlparser.Sentence{words, tags})
			words = make([]string, 0)
			tags = make([]string, 0)
		} else {
			parts := strings.Split(train.Text(), "\t")
			words = append(words, parts[0])
			tags = append(tags, parts[1])
		}
	}
	f.Close()
	if err := train.Err(); err != nil {
		log.Fatal(err)
	}

	p.Test(sentences)
}

func tag() {
	p := nlparser.New()

	fm, err := os.Open(*modelPath)
	if err != nil {
		log.Fatal(err)
	}
	dec := gob.NewDecoder(fm)
	model := &mlearning.Model{}
	if err := dec.Decode(model); err != nil {
		log.Fatal(err)
	}
	p.Perceptron.Model = model

	words := flag.Args()[1:]
	s := nlparser.Sentence{words, nil}
	p.Predict(&s)

	for i, w := range s.Words {
		log.Printf("%s - %s\n", w, s.Tags[i])
	}
}
