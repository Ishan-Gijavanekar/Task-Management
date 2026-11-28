APP_NAME := notification-service

.PHONY: run build fmt vet test test-race coverage tidy quality

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

tidy:
	go mod tidy

quality: fmt vet test build