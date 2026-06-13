.PHONY: all client server clean

all: client server

client:
	@echo "🦍 Compiling client binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-client ./cmd/client

server:
	@echo "🦍 Compiling server binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-server ./cmd/server

clean:
	rm -rf bin
