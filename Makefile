BINARY_NAME=qube
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
MAIN=.
BUILD_DIR=./bin

.PHONY: build clean

build:
	@echo "🔧 Building $(BINARY_NAME) version $(VERSION)"
	go build \
		-ldflags="-X github.com/apiqube/cli/cmd.version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION).exe $(MAIN)

clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)

go-fmt:
	gofumpt -l -w .

go-lint:
	golangci-lint run ./...