gen: 
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    server/pb/server.proto

build_all:
	go build -o ./bin/main main.go
run: build_all
	sudo ./bin/main


clean:
	rm ./bin/main

