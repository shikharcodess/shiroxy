# Default mode is 'dev'
MODE ?= dev

# URLs and flags for different modes
ifeq ($(MODE),dev)
    ACME_SERVER_URL=https://127.0.0.1:14000/dir
    INSECURE_SKIP_VERIFY=yes
endif
ifeq ($(MODE),stage)
    ACME_SERVER_URL=https://acme-staging-v02.api.letsencrypt.org/directory
    INSECURE_SKIP_VERIFY=yes
endif
ifeq ($(MODE),prod)
    ACME_SERVER_URL=https://acme-v02.api.letsencrypt.org/directory
    INSECURE_SKIP_VERIFY=no
endif

.PHONY: build test fmt lint all

build:
	@echo "Building in $(MODE) mode"
	@mkdir -p build
	# Ensure this line is tab-indented
	go build -ldflags "-X main.ACME_SERVER_URL=$(ACME_SERVER_URL) -X main.INSECURE_SKIP_VERIFY=$(INSECURE_SKIP_VERIFY)" -o build/shiroxy cmd/shiroxy/main.go

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run

all: fmt lint test build

# make MODE=dev/stage/prod