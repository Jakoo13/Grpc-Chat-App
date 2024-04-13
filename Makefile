build_app:
	docker build --tag=docker_example .

run_app:
	docker run -it -p 8080:8080 docker_example

protoAll:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	proto/*.proto