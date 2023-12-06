# Build container
FROM golang:1.21 AS build

ARG VERSION
ARG SOCKET NATS NATS_NKEY STREAM_PREFIX STREAM_NETWORK_INFIX

COPY . /home/src
WORKDIR /home/src

RUN CGO_ENABLED=0 go build -v -o . -installsuffix cgo -ldflags="-w -s" ./...

# Runtime container
FROM alpine

ENV PATH="/app:${PATH}"

RUN apk add --update --no-cache \
    ca-certificates \
    bash

WORKDIR /home/app
COPY --from=build /home/src/arbitrum-publisher .
COPY --from=build /home/src/docker/entrypoint.sh /home/app
RUN chmod +x /home/app/entrypoint.sh

ENTRYPOINT ["/home/app/entrypoint.sh"]
