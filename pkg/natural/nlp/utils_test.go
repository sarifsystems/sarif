// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import "strings"

func tagged(s string) []*Token {
	ts := make([]*Token, 0)
	for _, w := range strings.Split(s, " ") {
		p := strings.Split(w, ":")
		t := &Token{
			Value: p[0],
			Lemma: p[0],
			Tags:  make(map[string]struct{}),
		}
		for _, tag := range p[1:] {
			t.Tags[tag] = struct{}{}
		}
		ts = append(ts, t)
	}
	return ts
}
