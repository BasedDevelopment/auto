.PHONY: clean

all: auto auto-tools

# Build executable for Eve program
auto:
	go mod download
	go build --ldflags "-s -w" -o bin/auto ./cmd/auto/

auto-tools:
	go mod download
	go build --ldflags "-s -w" -o bin/auto-tools ./cmd/auto-tools/

test:
	go clean -testcache
	go test -v ./...

# Build and execute Eve program
start: auto
	./bin/auto --log-format pretty

# Format Sojourner source code with Go toolchain
format:
	go mod tidy
	go fmt ./...

# Clean up binary output folder
clean:
	rm -rf bin/
