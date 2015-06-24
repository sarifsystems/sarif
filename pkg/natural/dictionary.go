// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

type Dictionary struct {
	Tags map[string]string
}

func (d *Dictionary) Tag(tokens []*Token) map[string]int {
	counts := make(map[string]int)

	for _, t := range tokens {
		if tag, ok := d.Tags[t.Lemma]; ok {
			counts[tag]++
			t.Tags[tag] = struct{}{}
		}
	}

	return counts
}

func (d *Dictionary) TagSingle(t *Token) string {
	if tag, ok := d.Tags[t.Lemma]; ok {
		return tag
	}
	return ""
}
