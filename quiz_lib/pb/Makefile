all:
	mkdir -p ./gen
	protoc --proto_path=. --go_out=./gen --go_opt=paths=source_relative --go-grpc_out=./gen --go-grpc_opt=paths=source_relative quiz_api.proto const.proto
