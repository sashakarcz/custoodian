.PHONY: proto build test clean install

# Generate protobuf code
proto:
	PATH=$(PATH):$(HOME)/go/bin buf generate

# Build the custodian binary
build: proto
	go build -o bin/custodian ./cmd/custodian

# Install dependencies
deps:
	go mod download
	go install github.com/bufbuild/buf/cmd/buf@latest

# Run tests
test: proto
	go test -v ./...

# Clean generated files
clean:
	rm -rf bin/
	rm -f pkg/config/*.pb.go

# Install custodian locally
install: build
	sudo cp bin/custodian /usr/local/bin/

# Format code
fmt:
	go fmt ./...
	buf format -w

# Lint code
lint: proto
	go vet ./...
	buf lint

# Run all checks
check: fmt lint test

# Generate example configuration
example:
	./bin/custodian generate --template-dir templates/gcp --output examples/output examples/simple.textproto