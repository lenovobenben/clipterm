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
MACOS_INTEL_NAME := $(BINARY)-$(VERSION)-macos-intel
MACOS_APPLE_SILICON_NAME := $(BINARY)-$(VERSION)-macos-apple-silicon

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

.PHONY: package-macos
package-macos: package-macos-intel package-macos-apple-silicon

.PHONY: package-macos-intel
package-macos-intel:
	rm -rf $(DIST_DIR)/$(MACOS_INTEL_NAME) $(DIST_DIR)/$(MACOS_INTEL_NAME).tar.gz
	mkdir -p $(DIST_DIR)/$(MACOS_INTEL_NAME)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CGO_CFLAGS="-arch x86_64" CGO_LDFLAGS="-arch x86_64" go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(MACOS_INTEL_NAME)/$(BINARY) ./cmd/clipterm
	cp README.md LICENSE $(DIST_DIR)/$(MACOS_INTEL_NAME)/
	LANG=C LC_ALL=C tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(MACOS_INTEL_NAME).tar.gz $(MACOS_INTEL_NAME)

.PHONY: package-macos-apple-silicon
package-macos-apple-silicon:
	rm -rf $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME) $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME).tar.gz
	mkdir -p $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CGO_CFLAGS="-arch arm64" CGO_LDFLAGS="-arch arm64" go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME)/$(BINARY) ./cmd/clipterm
	cp README.md LICENSE $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME)/
	LANG=C LC_ALL=C tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(MACOS_APPLE_SILICON_NAME).tar.gz $(MACOS_APPLE_SILICON_NAME)

.PHONY: uninstall
uninstall:
	rm -f $(BINDIR)/$(BINARY)

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)
