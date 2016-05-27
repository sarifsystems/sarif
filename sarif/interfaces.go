// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

type writer interface {
	Write(msg Message) error
}

type reader interface {
	Read() (Message, error)
}

type Conn interface {
	Read() (Message, error)
	Write(msg Message) error
	Close() error
}

func transmitTo(r reader, w writer) error {
	for {
		msg, err := r.Read()
		if err != nil {
			return err
		}
		if err := w.Write(msg); err != nil {
			return err
		}
	}
}

func Transmit(a, b Conn) error {
	defer a.Close()
	defer b.Close()

	errs := make(chan error, 2)
	go func() {
		errs <- transmitTo(a, b)
	}()
	go func() {
		errs <- transmitTo(b, a)
	}()
	return <-errs
}
