// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

type Connection interface {
	Publish(msg Message) error
	Subscribe(src, action, device string) error
	Consume() (<-chan Message, error)
	Close() error
}
