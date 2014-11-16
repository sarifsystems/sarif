Stark Specification Version 0.4
=================================

A stark message is a JSON-encoded object with a number of required fields. This message
is normally send over TCP/TLS to a broker which dispatches it to the correct client(s).
There can also be other intermediate delivery methods such as JSON over WebSocket or
pure text over XMPP, but these are limited and non-standard.

Standard message fields
-----------------------

As defined in [proto.Message](http://godoc.org/github.com/xconstruct/stark/proto#Message).

* stark: semantic version of the protocol
* id: unique ID of this message (normally 8 alphanumeric chars)
* action: type of the message (delimited by '/', example: 'push/link')
* src: device ID of the sender (normally "name-id")
* dst: optional device ID of the destination
* p: optional payload (a JSON object with further fields)
* corr: optional correlation id of previous message
* text: optional user friendly textual representation of the message (recommended)

Default actions
---------------

Here is an overview of commonly used actions, their payload and expected
behaviour:

* ping: pings a device, expects an 'ack' as reply
* ack: affirmative reply
* proto/sub: subscribe to a specific action and/or device
	+ action: the action to subscribe to (matches the whole subtree).
	+ device: the device to subscribe to or empty for global broadcasts
* push/{content-type}: pushes content to the device, ready for consumption by
  the user
	+ content: the content string
	+ type: MIME type of the content
	+ title: optional title
