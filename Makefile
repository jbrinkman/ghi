# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=ghi
LOCAL_BIN=bin
GOPATH_BIN=$(shell go env GOPATH)/bin

# Version information (support both upper and lowercase variables)
VERSION ?= $(if $(version),$(version),dev)
DATE ?= $(if $(date),$(date),$(shell date -u +"%Y-%m-%d"))
COMMIT ?= $(if $(commit),$(commit),$(shell git rev-parse --short HEAD))
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

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

