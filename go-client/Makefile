.PHONY: proto basic-example

proto:
	protoc --go_out=./grpc/gen --go_opt=paths=source_relative,Mconfiguration.proto=github.com/necroskillz/config-service/go-client/grpc/gen \
		--go-grpc_out=./grpc/gen --go-grpc_opt=paths=source_relative,Mconfiguration.proto=github.com/necroskillz/config-service/go-client/grpc/gen \
		--proto_path=../backend/proto \
		configuration.proto

basic-example:
	go run ./example/basic/main.go

test:
	go test ./...

cover:
	go test -coverprofile=tmp/coverage.out ./...
	go tool cover -html=tmp/coverage.out