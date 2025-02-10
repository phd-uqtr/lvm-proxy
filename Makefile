gen: 
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    server/pb/server.proto

build_all:
	go build -o main main.go
run: build_all
	sudo ./main


clean:
	rm main

