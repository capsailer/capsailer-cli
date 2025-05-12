.PHONY: build clean test install lint release release-all help example-manifest example-build docker-build

BINARY_NAME := capsailer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%FT%T%z)
GOFLAGS := -trimpath
LDFLAGS := -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

# Detect OS
ifeq ($(OS),Windows_NT)
    BINARY_SUFFIX := .exe
    RM := del /f
    MKDIR := mkdir
else
    BINARY_SUFFIX :=
    RM := rm -f
    MKDIR := mkdir -p
endif

# Set the binary path
BIN_DIR := bin
BINARY_PATH := $(BIN_DIR)/$(BINARY_NAME)$(BINARY_SUFFIX)
SRC_FILES := cmd/capsailer/main.go cmd/capsailer/commands.go

# Build variables for release
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
RELEASE_DIR := release

# Default target
.DEFAULT_GOAL := build

# Display help information
help:
	@echo "Capsailer Makefile"
	@echo "Available targets:"
	@echo "  build        - Build the binary for the current platform"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run unit tests"
	@echo "  lint         - Run linting checks"
	@echo "  install      - Install binary to /usr/local/bin"
	@echo "  release      - Create release package for current platform"
	@echo "  release-all  - Create release packages for all supported platforms"
	@echo "  docker-build - Build the Docker image"
	@echo ""
	@echo "Example usage targets:"
	@echo "  example-manifest - Copy example manifest to manifest.yaml"
	@echo "  example-build    - Build a bundle from the example manifest"

# Create the bin directory
$(BIN_DIR):
	$(MKDIR) $(BIN_DIR)

# Build the binary
build: $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_PATH) $(SRC_FILES)
	@echo "Binary built at $(BINARY_PATH)"

# Clean build artifacts
clean:
	go clean
	$(RM) $(BINARY_NAME)
	$(RM) -r $(BIN_DIR)
	$(RM) -r $(RELEASE_DIR)
	$(RM) -r capsailer-build
	$(RM) *.tar.gz

# Run unit tests
test:
	go test -v ./...

# Run linting checks
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	$(HOME)/go/bin/golangci-lint run ./...

# Install the binary to /usr/local/bin
install: build
	cp $(BINARY_PATH) /usr/local/bin/
	@echo "Installed $(BINARY_NAME) to /usr/local/bin/"

# Build a release for the current platform
release: build
	@$(MKDIR) $(RELEASE_DIR)
	tar -czf $(RELEASE_DIR)/$(BINARY_NAME)-$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz $(BINARY_PATH)
	@echo "Release package created in $(RELEASE_DIR)/"

# Build releases for all platforms
release-all:
	@$(MKDIR) $(RELEASE_DIR)
	@echo "Building releases for all platforms..."
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		binary_name=$(BINARY_NAME); \
		if [ "$$os" = "windows" ]; then binary_name="$(BINARY_NAME).exe"; fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$$binary_name $(SRC_FILES); \
		if [ "$$os" = "windows" ]; then \
			cd $(RELEASE_DIR) && zip $(BINARY_NAME)-$(VERSION)-$$os-$$arch.zip $$binary_name && rm $$binary_name; \
		else \
			cd $(RELEASE_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-$$os-$$arch.tar.gz $$binary_name && rm $$binary_name; \
		fi; \
	done
	@echo "All release packages created in $(RELEASE_DIR)/"

# Example targets for convenience
example-manifest:
	cp examples/manifest.yaml manifest.yaml
	@echo "Copied example manifest to manifest.yaml"

example-build: build example-manifest
	$(BINARY_PATH) build --manifest manifest.yaml --output capsailer-bundle.tar.gz
	@echo "Bundle built as capsailer-bundle.tar.gz"
