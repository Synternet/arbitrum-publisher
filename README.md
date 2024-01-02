# Arbitrum publisher
[![Latest release](https://img.shields.io/github/v/release/SyntropyNet/Arbitrum-Publisher)](https://github.com/SyntropyNet/Arbitrum-Publisher/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/SyntropyNet/Arbitrum-Publisher/github-ci.yml?label=github-ci)](https://github.com/SyntropyNet/Arbitrum-Publisher/actions/workflows/github-ci.yml)

Establishes connection with Arbitrum node and publishes Arbitrum blockchain data to Syntropy Data Layer via NATS connection.

# Usage

Building from source.
```bash
make build
```

Running executable.
```bash
./arbitrum-publisher --socket /arbitrum.ipc --nats nats://35.198.163.186 --stream-prefix my-org --nats-nkey SA..BC
```

### Environment variables and flags {#env-flags}

Environment variables can be passed to docker container. Flags can be passed as executable arguments.

| Environment variable   | Flag                   | Description                                                                                                               |
| ---------------------- | ---------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| SOCKET                 | socket                 | Arbitrum node URI to establish IPC/WebSocket connection, e.g.: `/tmp/arbvitrum.ipc`, `ws://a.b.c.d:8546`                  |
| NATS                   | nats                   | NATS connection URL to Syntropy Data Layer broker, e.g.: `nats://e.f.g.h`. URL to [broker](https://docs.syntropynet.com/docs/actors/broker). Default value is set to testnet broker: `nats://35.198.163.186`.                                                                                                                      |
| NATS_NKEY              | nats-nkey              | NATS account NKEY a.k.a access token, e.g.: `SA..SI` (58 chars). One access token for each publisher, see [here](https://docs.syntropynet.com/build/data-layer/developer-portal/publish-streams#7-get-the-access-token).                                                                                                        |
| STREAM_PREFIX          | stream-prefix          | Stream prefix, e.g.: `foo` prefix results in `foo.arbitrum.<tx,log-even,header,...>` stream subjects. Stream prefix should be same as registered wallet [alias](https://docs.syntropynet.com/build/data-layer/developer-portal/publish-streams#2-register-a-wallet---get-your-alias).                                     |
| STREAM_PUBLISHER_INFIX | stream-publisher-infix | (optional) Stream publisher infix, e.g.: `foo` infix results in `prefix.foo.<tx,log-even,header,...>` stream subjects. Stream publisher infix should be same as registered publisher [alias](https://docs.syntropynet.com/build/data-layer/developer-portal/publish-streams#3-register-a-publisher). Default: `arbitrum`. |
| STREAM_NETWORK_INFIX   | stream-network-infix   | (optional) Specify stream network infix, e.g.: `mainnet` prefix results in `<prefix>.arbitrum.mainnet.<tx,...>` subjects. Default: empty (`prefix.arbitrum.<tx,...>`). |

See [Data Layer Quick Start](https://docs.syntropynet.com/build/data-layer/data-layer-quick-start) to learn more.

## Docker

### Build from source

1. Build image.
```
docker build -f ./docker/Dockerfile -t arbitrum-publisher .
```

2. Run container with passed environment variables.
```
docker run -it --rm --env-file=.env arbitrum-publisher
```

### Prebuilt image

Run container with passed environment variables.
```
docker run -it --rm --env-file=.env ghcr.io/syntropynet/arbitrum-publisher:latest
```

### Docker Compose

`docker-compose.yml` file.
```
version: '3.8'

services:
  arbitrum-publisher:
    image: ghcr.io/syntropynet/arbitrum-publisher:latest
    environment:
      - SOCKET=/tmp/arbitrum.ipc
      - NATS=nats://35.198.163.186
      - NATS_NKEY=secret-access-token
      - STREAM_PREFIX=some-org
      - STREAM_PUBLISHER_INFIX=arbitrum
    volumes:
      - /tmp/arbitrum.ipc:/tmp/arbitrum.ipc:ro
```

## Contributing

We welcome contributions from the community. Whether it's a bug report, a new feature, or a code fix, your input is valued and appreciated.

## Syntropy

If you have any questions, ideas, or simply want to connect with us, we encourage you to reach out through any of the following channels:

- **Discord**: Join our vibrant community on Discord at [https://discord.com/invite/jqZur5S3KZ](https://discord.com/invite/jqZur5S3KZ). Engage in discussions, seek assistance, and collaborate with like-minded individuals.
- **Telegram**: Connect with us on Telegram at [https://t.me/SyntropyNet](https://t.me/SyntropyNet). Stay updated with the latest news, announcements, and interact with our team members and community.
- **Email**: If you prefer email communication, feel free to reach out to us at devrel@syntropynet.com. We're here to address your inquiries, provide support, and explore collaboration opportunities.
