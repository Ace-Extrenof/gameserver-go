.PHONY: client server test

client:
	@go build -o bin/client client/main.go
	@./bin/client

server:
	@go build -o bin/server server/main.go
	@./bin/server

test:
	@go build -o bin/test test/main.go
	@./bin/test
