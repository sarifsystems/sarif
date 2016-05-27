#!/bin/bash

# This minimal sarif client connects to a broker at localhost:23100
# and pings the network.

netcat -t 3 localhost 23100 <<EOF
{
	"sarif":"0.5",
	"id": "$RANDOM",
	"action": "proto/sub",
	"src": "minimal.sh",
	"p": { "device": "minimal.sh" }
}

{
	"sarif":"0.5",
	"id": "$RANDOM",
	"action": "ping",
	"src": "minimal.sh"
}
EOF
