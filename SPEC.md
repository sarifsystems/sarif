Stark Specification Version 0.3
=================================

A stark message is a JSON-encoded object with a number of fixed fields. This message
is normally send over the [MQTT](http://mqtt.org) protocol to a broker which dispatches
it to the correct client(s). There can also be other intermediate delivery methods such
as JSON over WebSocket or pure text over XMPP, but these are limited and non-standard.

Standard message fields
-----------------------

As defined in [proto.Message](http://godoc.org/github.com/xconstruct/stark/proto#Message).

* v: semantic version of the protocol
* id: unique ID of this message (normally 8 alphanumeric chars)
* action: type of the message (delimited by '/', example: 'push/link')
* src: device ID of the sender (normally "name-id")
* dst: optional device ID of the destination
* p: optional payload (a JSON object with further fields)
* corr: optional correlation id of previous message

Default actions
---------------

Here is an overview of commonly used actions, their payload and expected
behaviour:

* ping: pings a device, expects an 'ack' as reply
* ack: affirmative reply
* push/{content-type}: pushes content to the device, ready for consumption by
  the user
	+ content: the content string
	+ type: MIME type of the content
	+ title: optional title

MQTT Topic mappings
-------------------

As defined in [proto.GetTopic](http://godoc.org/github.com/xconstruct/stark/proto#GetTopic).
Devices normally subscribe to a whole topic tree (e.g. "/stark/dev/mydevice/#" or "/stark/special/all/action/push/#").
This means that if you subscribe to an action "push", you also receive all subactions, e.g. "push/link" or "push/text".

* /stark/dev/{destination}: send message to a specific device
* /stark/special/all/{action}: send message to all devices
* /stark/user/{user}/{action}: send message to all devices with an active user in the last few minutes (planned)
