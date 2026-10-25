.PHONY: test vet contract generate types-check build

test:
	go test ./...

vet:
	go vet ./...

contract:
	go run ./cmd/contract-check

generate:
	go run ./cmd/openapi-types

types-check:
	go run ./cmd/openapi-types -check

build:
	go build ./cmd/server
