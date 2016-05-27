// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"fmt"
	"testing"
	"time"
)

func TestNet(t *testing.T) {
	msg := Message{
		Version:     VERSION,
		Id:          GenerateId(),
		Action:      "ping/something",
		Source:      "someone",
		Destination: "this",
	}

	l, err := Listen(&NetConfig{
		Address: "tcp://",
	})
	if err != nil {
		t.Fatal(err)
	}

	recv := make(chan Message)
	go func() {
		srv, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			for {
				msg, err := srv.Read()
				if err != nil {
					t.Fatal(err)
				}
				recv <- msg
			}
		}()
	}()

	fmt.Println(l.Addr())
	client, err := Dial(&NetConfig{
		Address: "tcp://" + l.Addr().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			_, err := client.Read()
			if err != nil {
				t.Fatal(err)
			}
		}
	}()
	if err := client.Write(msg); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-recv:
		t.Log(got)
		if got.Id != msg.Id {
			t.Fatal("wrong id:", got.Id, "expected:", msg.Id)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("no message received")
	}
}
