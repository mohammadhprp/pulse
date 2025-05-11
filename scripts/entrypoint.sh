#!/bin/sh

if [ "$0" = "./agent" ] || [ "$1" = "agent" ]; then
  exec ./bin/agent
else
  exec ./bin/collector
fi
