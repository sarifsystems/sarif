// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

func SplitQuoted(s string, sep string) ([]string, bool) {
	return genSplit(s, sep, 0)
}

func genSplit(s, sep string, sepSave int) ([]string, bool) {
	c := sep[0]
	start := 0

	a := make([]string, 0)
	quote := uint8(0)
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i] == quote {
			quote = 0
		} else if quote == 0 {
			if s[i] == '"' || s[i] == '`' {
				quote = s[i]
			} else if s[i] == c && (len(sep) == 1 || s[i:i+len(sep)] == sep) {
				a = append(a, s[start:i+sepSave])
				start = i + len(sep)
				i += len(sep) - 1
			}
		}
	}
	a = append(a, s[start:])
	return a, quote != 0
}

func TrimQuotes(s string) string {
	if s == "" {
		return s
	}
	q := s[0]
	if (q != '"' && q != '`') || s[len(s)-1] != q {
		return s
	}
	return s[1 : len(s)-1]
}
