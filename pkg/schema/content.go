// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

type Content struct {
	*Thing
	Url       string `json:"url,omitempty"`
	PutAction string `json:"put_action,omitempty"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`

	Data []byte `json:"-"`
}

func (c Content) HasData() bool {
	return len(c.Data) > 0
}
