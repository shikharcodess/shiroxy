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

`protoc --go_out=. ./cmd/shiroxy/storage/storage.proto`
