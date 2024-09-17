proto:
	protoc pkg/pb/*.proto --go_out=.

proto_grpc:
	protoc pkg/pb/*.proto --go-grpc_out=require_unimplemented_servers=false:.

server:
	go run cmd/server.go