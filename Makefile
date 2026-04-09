BINARY_NAME ?= taskmd
VERSION ?= dev
DIST_DIR ?= dist
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/ferdikt/taskmd-cli/internal/buildinfo.Version=$(VERSION) \
	-X github.com/ferdikt/taskmd-cli/internal/buildinfo.Commit=$(COMMIT) \
	-X github.com/ferdikt/taskmd-cli/internal/buildinfo.Date=$(DATE)

.PHONY: build run tidy fmt test build-all checksums dist brew-dist clean

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

run:
	go run .

tidy:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

test:
	go test ./...

build-all: clean
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64 .

checksums:
	cd $(DIST_DIR) && shasum -a 256 * > checksums.txt

dist: build-all checksums
	@echo "dist artifacts ready in $(DIST_DIR)"

brew-dist: clean
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "brew-dist artifacts ready in $(DIST_DIR)"

clean:
	rm -rf bin $(DIST_DIR)

