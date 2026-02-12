
BINARY_NAME=paqet

.PHONY: all build build-linux clean install

all: build

# Build for the current OS
build:
	go build -o $(BINARY_NAME) ./cmd

# Cross-compile for Linux (useful for deploying to VPS from Mac/Windows)
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux ./cmd

# Install to /usr/local/bin (Mac/Linux)
install: build
	mv $(BINARY_NAME) /usr/local/bin/

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux
