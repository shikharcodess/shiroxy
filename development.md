# Shiroxy Development Guide:

## 1. Setup Pebble

- <strong>Clone the Pebble Repository</strong>

```bash
git clone https://github.com/letsencrypt/pebble.git
```

- <strong>Build Pebble</strong>

```bash
go build -o pebble  ./cmd/pebble/main.go
```

## 2. Clone Repo and Run

- <strong>Clone the shiroxy Repository</strong>

```bash
git clone https://github.com/ShikharY10/shiroxy
```

- <strong>Run shiroxy as Reverse Proxy

```bash
cd /shiroxy
go run /cmd/main.go -c shiroxy.config.json
```

# Development Phase Sprints

### Sprint 1 Tasks: [Completed (13 April 2024)]

- Configuration Reader
- Logger
- Loader
- Logo
- DNS Challenge Solver Setup

### Sprint 2 Tasks [Due Date: 31 May 2024 ]

- Proxy Frontend Server
- Proxy Backend Server
- Implement Support of TLS for frontend and backend servers
- Implement DNS challenge Solver

### Sprint 3 Taks [Due Date: 10 June 2024]

- Redis Integration for storaging for cert data
- Implement Inhouse Implementation for storing cert data in primary memory
- Analytics API
- Analytics

### Sprint 4 Tasks [Due Date: 30 June 2024]

- Health Check
- Webhook
- RestAPI for dynamic registering and removing domains
- Remote Logger
- Remote Configuration Manager

# Compiling Proto file

`protoc --go_out=. ./cmd/shiroxy/storage/storage.proto`
