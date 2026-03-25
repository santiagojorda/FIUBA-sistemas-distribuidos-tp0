#!/usr/bin/env bash

IP_SUBNET="172.25.125.0/24"
MAX_BATCH_AMOUNT="${MAX_BATCH_AMOUNT:-120}"
MAX_BATCH_SIZE="${MAX_BATCH_SIZE:-8192}"

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <OUTPUT_FILE> <AMOUNT_CLIENTS>"
    exit 1
fi

OUTPUT_FILE=$1
AMOUNT_CLIENTS=$2

if ! [[ $AMOUNT_CLIENTS =~ ^[0-9][0-9]*$ ]]; then
    echo "Error: Number of clients must be a positive integer."
    exit 1
fi

# Server 
cat > $OUTPUT_FILE <<EOL
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: ["python3", "/server/main.py"]
    environment:
      - PYTHONUNBUFFERED=1
      - AMOUNT_CLIENTS=$AMOUNT_CLIENTS
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
EOL

# Clients
for i in $(seq 1 $AMOUNT_CLIENTS); do
  cat >> $OUTPUT_FILE <<EOL
  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: ["/client"]
    restart: no
    environment:
      - CLI_ID=$i
      - CLI_BATCH_MAXAMOUNT=$MAX_BATCH_AMOUNT
      - CLI_BATCH_MAXSIZE=$MAX_BATCH_SIZE
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data:/.data:ro
EOL

done

# Red
cat >> $OUTPUT_FILE <<EOL
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: $IP_SUBNET
EOL