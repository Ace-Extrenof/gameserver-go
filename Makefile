.PHONY: client server test

client:
	@go build -o bin/client client/main.go
	@./bin/client

server:
	@go build -o bin/server server/main.go
	@./bin/server

test:
	@cd client; python3 -m http.server 8001
