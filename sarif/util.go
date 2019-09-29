// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"crypto/rand"
	"errors"
	"net/url"
)

func GenerateId() string {
	const alphanum = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var bytes = make([]byte, 8)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func ConvertURL(u string) (Message, error) {
	msg := Message{}
	us, err := url.Parse(u)
	if err != nil {
		return msg, err
	}
	if us.Scheme != "sarif" {
		return msg, errors.New("Expected url scheme sarif://")
	}

	msg.Action = us.Host + us.Path
	if us.User != nil {
		msg.Destination = us.User.Username()
	}
	err = msg.EncodePayload(us.Query())
	return msg, err
}
