// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

func BadRequest(reason error) Message {
	str := "Bad Request"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "err/badrequest",
		Text:   str,
	}
}

func InternalError(reason error) Message {
	str := "Internal Error"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "err/internal",
		Text:   str,
	}
}
