# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=ghi
LOCAL_BIN=bin
GOPATH_BIN=$(shell go env GOPATH)/bin

# Version information
VERSION ?= dev
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%d")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"

.PHONY: all build test clean tag install

all: test build

build: 
	mkdir -p $(LOCAL_BIN)
	$(GOBUILD) $(LDFLAGS) -o $(LOCAL_BIN)/$(BINARY_NAME)

install:
	mkdir -p $(GOPATH_BIN)
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH_BIN)/$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

clean:
	rm -rf $(LOCAL_BIN)
	rm -f $(GOPATH_BIN)/$(BINARY_NAME)

# Create a new version tag
tag:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Please specify VERSION=X.X.X when running make tag"; \
		exit 1; \
	fi
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)

