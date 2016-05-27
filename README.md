sarif
=====

[![Build](https://img.shields.io/travis/sarifsystems/sarif.svg?style=flat-square)](https://travis-ci.org/sarifsystems/sarif)
[![API Documentation](https://img.shields.io/badge/api-GoDoc-blue.svg?style=flat-square)](https://godoc.org/github.com/sarifsystems/sarif)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](http://opensource.org/licenses/MIT)

Sarif is a simple network of microservices to perform tasks across devices for the user.
Combined with APIs and user-tracking, it serves as a personal assistant.
This is the message spec and reference implementation, written in Go.

Since it is currently in a prototype stage, it has a limited set of features out of
the box and there may be a lot of breaking changes. There is currently no documentation.
Here be dragons. That said, the core functionality is there.

## In Detail ##

Sarif aspires to be a personal helper that has access to a range of different
tools to aid in automating everyday life, an "intranet of apps".  For example,
your phone tracks your location and when Sarif notices that you are coming home,
it boots your desktop pc for you and starts the music. Or reacts to chat
commands, or displays notifications on your watch.

A microservice could be anything, e.g. a location publisher on your phone,
a database server, a media player control, a webservice, a voice assistant / chatbot,
or your personal context-aware artifical intelligence robot overlord.

Sarif provides:

* A very simple message format based on JSON (see [SPEC.md](SPEC.md))
* A network of extensible, interconnected services across devices
* A natural user interface.
* A data platform for quantified self tracking and context-awareness
* Hopefully in the future something that can be called an AI

## Getting Started

	$ go install github.com/sarifsystems/sarif/cmd/sarifd
	$ go install github.com/sarifsystems/sarif/cmd/tars

	$ sarifd -v
	$ tars
	> .ping
	> remind me in 10 seconds that this thing works
	> .full
	> .cmd/catfacts

And take a look at `cmd/examples/sarifping`.

## Design Goals ##

**Interface-agnostic**: Sarif should work "everywhere" and degrade gracefully. If
you are on your phone, it should react to text messages. If you use your computer,
it should display a dashboard, system notifications and rich controls. If you are
at the commandline, Sarif should provide scripting and easy data access. If you are at
home, a voice assistant could listen for commands. All these can be implemented as
separate services.

**Simple:** Adding services should be as easy and future-proof as possible.
Your service should not depend on a big Sarif library, a specific programming
language or a complicated message format. That is why a Sarif message is a simple
JSON object sent over TCP/TLS. Ideally, you could connect your application to
the Sarif network in under 100 lines of code, from scratch without any libraries.

**Modular**: There should be no overarching core service that handles everything,
all services should be exchangeable and stand on equal footing. For example,
it should be possible to remove the NLP service that understands your text commands
and replace it with your own voice-controlled robot. Hey, you could even take
the message specification and write your own server.

Inspired by:

* Google Now / Siri / Cortana
* The Internet of Things
* If This Then That
* JARVIS, TARS, GERTY, Samaritan and Skynet
