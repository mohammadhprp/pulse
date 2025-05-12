#!/bin/sh

if [ "$1" = "agent" ] || [ "$2" = "agent" ]; then
  echo "Starting agent..."
  exec ./bin/agent
else
  echo "Starting collector..."
  exec ./bin/collector
fi
