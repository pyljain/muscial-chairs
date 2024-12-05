build-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./internal/proto/worker.proto

build: build-proto
	go build -o mc

test:
	curl -X POST -d '{ "worker_id": "mc-worker-pleased-arachnid" }' http://localhost:8080/busy