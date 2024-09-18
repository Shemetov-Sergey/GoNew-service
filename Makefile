proto:
	protoc --proto_path=pkg/pb/ --go_out=pkg/pb/ --go_opt=paths=source_relative pkg/pb/**/*.proto

proto_grpc:
	protoc --proto_path=pkg/pb/ --go-grpc_out=require_unimplemented_servers=false:pkg/pb/ --go-grpc_opt=paths=source_relative pkg/pb/**/*.proto

server:
	go run cmd/server.go