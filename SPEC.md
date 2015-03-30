Stark Specification Version 0.5
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
* src: device ID of the sender (normally "host/name/id")
* dst: optional device ID of the destination
* p: optional payload (a JSON object with further fields)
* corr: optional correlation id of previous message for replies/triggered messages
* text: optional user friendly textual representation of the message (recommended)

Example conversation
--------------------

Subscribe to all messages directed at our device with name "mydevice/123".

```json
{
	"stark":"0.5",
	"id": "1234abcd",
	"action": "proto/sub",
	"src": "mydevice/123",
	"p": { "action": "", "device": "mydevice/123" }
}
```

Publish a `ping` message that prompts all connected devices to respond.

```json
{
	"stark":"0.5",
	"id": "789xyz",
	"action": "ping",
	"src": "mydevice/123"
}
```

Or ping a specific device.

```json
{
	"stark":"0.5",
	"id": "789xyz",
	"action": "ping",
	"src": "mydevice/123",
	"dst": "phone/dog"
}
```

A possible reply to our `ping` message.

```json
{
	"stark":"0.5",
	"id": "asdas125",
	"action": "ack",
	"src": "phone/dog",
	"dst": "mydevice/123",
	"corr": "789xyz"
}
```

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
