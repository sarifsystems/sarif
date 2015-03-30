#!/bin/bash

# This minimal stark client connects to a broker at localhost:23100
# and pings the network.

netcat -t 3 localhost 23100 <<EOF
{
	"stark":"0.5",
	"id": "$RANDOM",
	"action": "proto/sub",
	"src": "minimal.sh",
	"p": { "device": "minimal.sh" }
}

{
	"stark":"0.5",
	"id": "$RANDOM",
	"action": "ping",
	"src": "minimal.sh"
}
EOF
