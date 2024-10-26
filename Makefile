.PHONY: client server

client:
	@go build -o bin/client client/main.go
	@./bin/client

server:
	@go build -o bin/server server/main.go
	@./bin/server
