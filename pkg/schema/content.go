// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

type Content struct {
	*Thing
	Url  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`

	Data []byte `json:"data,omitempty"`
}

func (c Content) HasData() bool {
	return len(c.Data) > 0
}
