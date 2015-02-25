// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

type Writer interface {
	Write(msg Message) error
}

type Reader interface {
	Read() (Message, error)
}

type Conn interface {
	Reader
	Writer
	Close() error
}

func TransmitTo(r Reader, w Writer) error {
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
		errs <- TransmitTo(a, b)
	}()
	go func() {
		errs <- TransmitTo(b, a)
	}()
	return <-errs
}
