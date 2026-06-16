.PHONY: all deamon dash server clean

all: deamon dash server

daemon:
	@echo "🦍 Compiling background client daemon binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-client-daemon ./cmd/client-daemon

dash:
	@echo "📟 Compiling client terminal dashboard tool..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-client-dash ./cmd/client-dash

server:
	@echo "🦍 Compiling server binary..."
	@mkdir -p bin
	go build -o bin/dumb-proxy-server ./cmd/server

clean:
	rm -rf bin
