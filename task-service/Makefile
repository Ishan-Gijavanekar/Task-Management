.PHONY: test vet generate
test:
	go test -race ./...
vet:
	go vet ./...
generate:
	protoc --go_out=. --go_opt=module=github.com/Ishan-Gijavanekar/task-service --go-grpc_out=. --go-grpc_opt=module=github.com/Ishan-Gijavanekar/task-service api/proto/task/v1/task.proto api/proto/task/v1/administration.proto
