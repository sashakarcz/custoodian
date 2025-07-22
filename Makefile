.PHONY: proto build test clean install

# Generate protobuf code
proto:
	PATH=$(PATH):$(HOME)/go/bin buf generate
	@go run scripts/move-proto.go

# Build the custoodian binary
build: proto
	@mkdir -p bin
ifeq ($(OS),Windows_NT)
	go build -o bin/custoodian.exe ./cmd/custoodian
else
	go build -o bin/custoodian ./cmd/custoodian
endif

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

# Install custoodian locally
install: build
	sudo cp bin/custoodian /usr/local/bin/

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
	./bin/custoodian generate --template-dir templates/gcp --output examples/output examples/simple.textproto