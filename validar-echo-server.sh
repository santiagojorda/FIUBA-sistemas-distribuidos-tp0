#!/usr/bin/env bash

MESSAGE="Testing echo server"
SUCCESS_MESSAGE="action: test_echo_server | result: success"
FAIL_MESSAGE="action: test_echo_server | result: fail"

PORT=12345
HOST="server"
NETWORK="tp0_testing_net"

RESP="$(docker run --rm --network "$NETWORK" busybox sh -c "printf '%s\n' '$MESSAGE' | nc -w 2 $HOST $PORT")"

if [ "$RESP" = "$MESSAGE" ]; then
  echo "$SUCCESS_MESSAGE"
else
  echo "$FAIL_MESSAGE"
fi