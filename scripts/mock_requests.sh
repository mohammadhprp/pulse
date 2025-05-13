#!/bin/bash

for i in {1..1000}
do
  timestamp=$(date +%s%3N)
  service="my-service"
  level="INFO"
  message="User $((RANDOM % 1000)) logged in"
  host="server-$((1 + RANDOM % 5))"

  json=$(jq -n \
    --arg time "$timestamp" \
    --arg service "$service" \
    --arg level "$level" \
    --arg message "$message" \
    --arg host "$host" \
    '{
      event_time_ms: ($time | tonumber),
      service: $service,
      level: $level,
      message: $message,
      host: $host
    }')

  echo "Sending event number: $i"
  curl -s -X POST http://localhost:8080/events \
    -H "Content-Type: application/json" \
    -d "$json" > /dev/null
done
