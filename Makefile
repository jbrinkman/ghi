# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=ghi

# Version information
VERSION ?= dev
GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%d")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"

.PHONY: all build test clean tag

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Create a new version tag
tag:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Please specify VERSION=X.X.X when running make tag"; \
		exit 1; \
	fi
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)

