# Contributing to shiroxy

Thank you for considering contributing to shiroxy! We appreciate your support and are excited to work with you. The following guidelines will help you understand how to contribute effectively.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on our GitHub repository. Include as much detail as possible:

- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots, if applicable

### Suggesting Features

We welcome feature suggestions! Please open an issue with the following details:

- A clear and concise description of the feature
- The problem it solves or the enhancement it provides
- Any relevant examples or mockups

### Submitting Pull Requests

1. **Fork the repository**
2. **Create a branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes**
4. **Commit your changes**: `git commit -m 'Add new feature'`
5. **Push to the branch**: `git push origin feature/your-feature-name`
6. **Open a pull request**: Go to our repository on GitHub and open a pull request.

### Development Setup

#### Prerequisites

- [Git](https://git-scm.com/)
- [Go](https://golang.org/)
- [Protoc](https://grpc.io/docs/protoc-installation)
- [Peblle](https://github.com/letsencrypt/pebble) (If you want to work with certification generation)

#### Setup Pebble

Clone the pebble repository

```bash
git clone https://github.com/letsencrypt/pebble.git
cd pebble
go build -o pebble  ./cmd/pebble/main.go

```

#### Installation

Clone the repository:

```bash
git clone https://github.com/yourusername/shiroxy.git
cd shiroxy
```
