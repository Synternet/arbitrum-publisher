#!/bin/sh

CMD="./arbitrum-publisher"

if [ ! -z "$SOCKET" ]; then
  CMD="$CMD --socket $SOCKET"
fi

if [ ! -z "$NATS" ]; then
  CMD="$CMD --nats $NATS"
fi

if [ ! -z "$NATS_NKEY" ]; then
  CMD="$CMD --nats-nkey $NATS_NKEY"
fi

if [ ! -z "$STREAM_PREFIX" ]; then
  CMD="$CMD --stream-prefix $STREAM_PREFIX"
fi

if [ ! -z "$STREAM_NETWORK_INFIX" ]; then
  CMD="$CMD --stream-network-infix $STREAM_NETWORK_INFIX"
fi

exec $CMD
