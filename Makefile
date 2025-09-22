VERSION=$(shell git describe --tags --always)
BINARY_NAME=ottermq
BUILD_DIR=bin

build: 
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.version=$(VERSION)" -o ./$(BUILD_DIR)/$(BINARY_NAME) ./cmd/ottermq/main.go


docs:
	@$(shell go env GOPATH)/bin/swag init --parseInternal  -g ../../../cmd/ottermq/main.go --pd -d web/handlers/api,web/handlers/api_admin -exclude web/handlers/webui/ -o ./web/static/docs -ot json

install:
	@mkdir -p $(shell go env GOPATH)/bin
	@mv ./$(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: build install clean run docs
