<div align="center">

<img src="https://raw.githubusercontent.com/shikharcodess/shiroxy/main/media/shiroxy_logo_.png" alt="shiroxy Logo">

# Shiroxy: Secure and Dynamic Web Traffic Management

**A Go-based reverse proxy designed for dynamic routing, SSL automation, and scalable domain management.**

</div>

## Quick Start

Get Shiroxy up and running in just a few minutes:

```bash
git clone https://github.com/shikharcodess/shiroxy.git
cd shiroxy
docker compose up -d --build
```

## Key Features

- **Automatic SSL Certificates**: Secure your domains effortlessly.
- **Custom Traffic Routing**: Tailor routing logic to fit your needs.
- **Dynamic Domain Management**: Manage domains flexibly via REST API.
- **System and Process Analytics**: Gain valuable insights with detailed analytics.

## Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.15+
- Docker
- Pebble (for local SSL testing)

## Installation

### Local Development

1. Clone the repository:

   ```bash
   git clone https://github.com/shikharcodess/shiroxy.git
   cd shiroxy
   ```

2. Run Pebble, the local test ACME server:

   ```bash
   cd pebble
   go build -o pebble cmd/pebble/main.go
   ./pebble
   ```

3. Start Shiroxy in development mode:
   ```bash
   sudo go run cmd/shiroxy/main.go -c /defaults/shiroxy.conf.yaml
   ```

### Using Docker

For a dockerized setup, run:

```bash
git clone https://github.com/shikharcodess/shiroxy.git
cd shiroxy
docker compose up -d --build
```

## Documentation

- [Configuration File](https://github.com/shikharcodess/shiroxy/blob/main/configuration.md)
- [Contribution Guide](https://github.com/shikharcodess/shiroxy/blob/main/CONTRIBUTION.md)
- [Interactive API Docs](https://github.com/shikharcodess/shiroxy/blob/main/docs/api.md)
- [License](https://github.com/shikharcodess/shiroxy/blob/main/LICENSE)

## Community and Support

Join our growing community:

- [Discussion Forum](https://github.com/shikharcodess/shiroxy/discussions/1)

## How to Contribute

Interested in contributing? Check out the [contribution guidelines](https://github.com/shikharcodess/shiroxy/blob/main/CONTRIBUTION.md) for more information on how you can contribute to Shiroxy.

## License

Shiroxy is MIT licensed. See the [LICENSE](https://github.com/shikharcodess/shiroxy/blob/main/LICENSE) file for more details.
