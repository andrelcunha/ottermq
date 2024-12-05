VERSION := $(shell git describe --tags --always)
BINARY_NAME=ottermq
CLIENT_BINARY_NAME=ottermq-client
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X got_it/cmd.version=$(VERSION)" -o ./$(BUILD_DIR)/$(BINARY_NAME) ./cmd/broker/main.go
	@go build -o ./$(BUILD_DIR)/$(CLIENT_BINARY_NAME) ./cmd/cli/main.go

install:
	@mkdir -p $(shell go env GOPATH)/bin
	@mv ./$(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@mv ./$(BUILD_DIR)/$(CLIENT_BINARY_NAME) $(shell go env GOPATH)/bin/$(CLIENT_BINARY_NAME)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@rm -f $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(CLIENT_BINARY_NAME)

.PHONY: build install clean run
