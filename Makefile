build:
    go build -o myapp cmd/myapp/main.go

test:
    go test ./...

fmt:
    gofmt -w .

lint:
    golangci-lint run

all: fmt lint test build
