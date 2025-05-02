.PHONY: generate test run

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/subpub.proto

test:
	go test -v ./internal/subpub/...

run:
	go run cmd/server/main.go