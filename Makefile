BUILD_DIR ?= out
BINARY_NAME ?= action
GO_BUILDFLAGS ?= -buildvcs=false -trimpath
LDFLAGS ?= -s -w -X main.version=$(shell cat VERSION 2>/dev/null || echo "dev")

.PHONY: all
all: build

.PHONY: build
build: $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/action

.PHONY: cross
cross: $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64    ./cmd/action
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64    ./cmd/action
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64   ./cmd/action
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GO_BUILDFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64   ./cmd/action

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor: tidy
	go mod vendor

.PHONY: test
test:
	go test ./...
