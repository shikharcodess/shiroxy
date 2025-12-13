[![codecov](https://codecov.io/github/shikharcodess/shiroxy/graph/badge.svg?token=6bVHb5fRuz)](https://codecov.io/github/shikharcodess/shiroxy)

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

- **Automatic SSL Certificates**: Secure your domains effortlessly with ACME protocol integration.
- **HTTP/2 Support**: Optimized HTTP/2 transport with connection pooling and multiplexing.
- **Intelligent Compression**: Automatic gzip compression for text-based content types.
- **Advanced Load Balancing**: Multiple strategies including round-robin, least-connection, and sticky-session.
- **Custom Traffic Routing**: Tailor routing logic with tag-based routing and caching.
- **Dynamic Domain Management**: Manage domains flexibly via REST API.
- **Performance Optimized**: Buffer pooling and connection reuse to minimize latency and resource usage.
- **System and Process Analytics**: Real-time monitoring of connections, request duration, and server health.
- **Graceful Shutdown**: Data persistence and clean shutdown with no request loss.

## Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.22+ (for latest HTTP/2 and performance features)
- Docker (optional, for containerized deployment)
- Pebble (for local SSL testing in development)

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

## What's New in v1.1.0 (Kuchii Release)

🚀 **Performance & Reliability Enhancements:**

- **HTTP/2 Support**: Full HTTP/2 implementation with connection pooling and multiplexing
- **Smart Compression**: Automatic gzip compression for compressible content (text, JSON, XML, etc.)
- **Buffer Pooling**: Efficient memory reuse with 32KB buffer pools to reduce GC pressure
- **Connection Statistics**: Real-time monitoring of connection pools, reuse rates, and request duration
- **Bug Fixes**:
  - Fixed buffer pool pointer issues preventing memory corruption
  - Corrected idle connection counter to prevent negative values
  - Fixed HTTP to HTTPS redirect URLs (removed port numbers)
  - Improved HTTP status codes (503 for inactive domains instead of 404)
  - Optimized header management for gzip compression

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
