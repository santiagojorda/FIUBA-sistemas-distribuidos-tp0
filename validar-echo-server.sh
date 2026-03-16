#!/usr/bin/env bash

MESSAGE="Testing echo server"
SUCCESS_MESSAGE="action: test_echo_server | result: success"
FAIL_MESSAGE="action: test_echo_server | result: fail"

PORT=12345
HOST="server"
NETWORK="tp0_testing_net"

if ! docker ps | grep -q "server"; then
  echo "Server container is not running."
  echo "$FAIL_MESSAGE"
  exit 1
fi

RESP="$(docker run --rm --network "$NETWORK" busybox sh -c "printf '%s\n' '$MESSAGE' | nc -w 2 $HOST $PORT" 2>/dev/null || true)"
RESP="$(printf '%s' "$RESP" | tr -d '\r\n')"

if [ "$RESP" = "$MESSAGE" ]; then
  echo "$SUCCESS_MESSAGE"
else
  echo "$FAIL_MESSAGE"
fi