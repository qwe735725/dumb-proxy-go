.PHONY: all client client-dash server clean

all: client client-dash server

client:
	@echo "🦍 Compiling client daemon binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-client-daemon ./cmd/client-daemon

client-dash:
	@echo "🦍 Compiling client dashboard tool..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-client-dash ./cmd/client-dash

server:
	@echo "🦍 Compiling server binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-server ./cmd/server

clean:
	rm -r bin
