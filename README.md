stark
=====

Stark is a simple communication protocol to perform tasks across devices for the user.
This is the reference implementation and spec, written in Go. Take a look at the
[blog post](http://xconstruct.net/2013/02/26/a-new-project-stark/) to see the
motivation and how it works.

Since it is currently in a prototype stage to develop the protocol, it has a limited set of features out of
the box and there may be a lot of breaking changes. No installation guide
is currently provided since it is nowhere near usable state, but if you are interested,
you can take a look at the "apps" folder to get started.

## In Detail ##

Stark aspires to be a personal helper that has access to a range of different
tools to aid in automating everyday life, an "intranet of apps".

* You search for an address on your desktop pc and want to navigate to it on your phone.
* You want to control the music in your room while lying in the bed with your tablet.
* Your phone notifies your router that you are coming home and boots your desktop pc for you.

Stark does not know how to do any of these things. But there may be an existing app or
cloud service that already knows this, but has an obscure API. Instead of writing
a commandline client to interact with it, you could connect it to stark and make
it part of your personal message bus. Another part in your growing army of tools,
accessible from any device connected with stark.

Stark provides:

* A decentralized network of interconnected services
* Support across multiple protocols and devices
* A natural user interface 
* A very simple and extendable message format based on JSON

Inspired by:

* Android Intents
* If This Then That (IFTTT)
* The Internet of Things
* JARVIS

## Use Case Example ##

1. You want to push a link from your desktop pc to your smartphone. You open a
   launcher on your desktop that connects to the stark network and enter: `push http://google.com to sgs2`
2. The launcher transmits your natural command to your desktop server.
4. The "natural" service on your desktop recognizes that you want to perform a `push`
   task with the link to a device named "sgs2".
5. The desktop server connects to the smartphone and transmits the `push` task.
6. The smartphone receives the `push` task and the notification services displays a notification with the link.

## FAQ ##

> Is it named after the character, the house or the world Stark?

Yes.
