VERSION=$(shell git describe --tags --always)
BINARY_NAME=ottermq
BUILD_DIR=bin
MAIN_PATH=cmd/ottermq/main.go

build: 
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.VERSION=$(VERSION)" -o ./$(BUILD_DIR)/$(BINARY_NAME) ./${MAIN_PATH}

docs:
	@$(shell go env GOPATH)/bin/swag init -g ../../../${MAIN_PATH} --pd -d web/handlers/api,web/handlers/api_admin -exclude web/handlers/webui/ -o ./web/docs --ot go

install:
	@mkdir -p $(shell go env GOPATH)/bin
	@mv ./$(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@go test ./... -v

lint:
	@golangci-lint run

ui-deps:
	@cd ottermq_ui && npm install

ui-build: ui-deps
	@cd ottermq_ui && npx quasar build
	@rm -rf ui
	@cp -r ottermq_ui/dist/spa ui

build-all: ui-build build

run-dev: build
	@echo "UI running separately at http://localhost:9000"
	@echo "Starting broker..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -rf ./ui
	@rm -rf ./ottermq_ui/dist

.PHONY: build install clean run docs test lint ui-deps ui-build build-all run-dev