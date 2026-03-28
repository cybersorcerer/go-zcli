BINARY_NAME := zcli
BINARY_DIR  := bin
INSTALL_DIR := $(HOME)/bin

VERSION := v0.4.0
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: build build-local build-zos run install clean dep vet lint test

build:
	@mkdir -p $(BINARY_DIR)
	@GOARCH=amd64 GOOS=darwin go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOARCH=arm64 GOOS=darwin go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@GOARCH=amd64 GOOS=linux  go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOARCH=s390x GOOS=linux  go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME)-linux-s390x .
	@GOARCH=amd64 GOOS=windows go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME)-windows.exe .

build-zos:
	@echo "Cross-compilation for z/OS is not supported."
	@echo "Build natively on z/OS with: go build -ldflags '$(LDFLAGS)' -o $(BINARY_NAME) ."

build-local:
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags '$(LDFLAGS)' -o $(BINARY_DIR)/$(BINARY_NAME) .

run: build-local
	./$(BINARY_DIR)/$(BINARY_NAME) --help

install: build-local
	@mkdir -p $(INSTALL_DIR)
	cp $(BINARY_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"

clean:
	go clean
	rm -rf $(BINARY_DIR)

dep:
	go mod download
	go mod tidy

vet:
	go vet ./...

lint:
	golangci-lint run --enable-all

test:
	go test ./...
