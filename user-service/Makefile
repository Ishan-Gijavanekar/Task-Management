GOCACHE ?= ../.gocache
COVER_PACKAGES := ./internal/auth ./internal/config ./internal/repository ./internal/service ./internal/transport/grpc ./internal/transport/grpc/interceptor ./internal/transport/http ./tests

.PHONY: fmt vet test test-race test-cover test-integration quality

fmt:
	go fmt ./...

vet:
	GOCACHE=$(GOCACHE) go vet ./...

test:
	GOCACHE=$(GOCACHE) go test ./...

test-race:
	GOCACHE=$(GOCACHE) go test -race ./...

test-cover:
	GOCACHE=$(GOCACHE) go test -coverprofile=coverage.out $(COVER_PACKAGES)
	go tool cover -func=coverage.out

test-integration:
	go test -count=1 -run Integration ./internal/repository/...

quality: fmt vet test
