BINARY := clipterm
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
VERSION ?= dev
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
DIST_DIR := dist
PACKAGE_DIR := $(DIST_DIR)/$(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH)
PACKAGE := $(PACKAGE_DIR).tar.gz
LDFLAGS := -X github.com/lenovobenben/clipterm/internal/version.Version=$(VERSION)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/clipterm

.PHONY: test
test:
	go test ./...

.PHONY: install
install: build
	mkdir -p $(BINDIR)
	cp $(BINARY) $(BINDIR)/$(BINARY)

.PHONY: package
package: build
	rm -rf $(PACKAGE_DIR) $(PACKAGE)
	mkdir -p $(PACKAGE_DIR)
	cp $(BINARY) $(PACKAGE_DIR)/$(BINARY)
	cp README.md LICENSE $(PACKAGE_DIR)/
	LANG=C LC_ALL=C tar -C $(DIST_DIR) -czf $(PACKAGE) $(notdir $(PACKAGE_DIR))

.PHONY: uninstall
uninstall:
	rm -f $(BINDIR)/$(BINARY)

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
