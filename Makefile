.PHONY: build install test clean release

BINARY_NAME=summon-wpm
VERSION=$(shell grep "version = " cmd/summon-wpm/main.go | cut -d '"' -f2)
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64
MAIN_PATH=./cmd/summon-wpm

build:
	go build -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)

install: build
	mkdir -p /usr/local/lib/summon
	cp $(BINARY_NAME) /usr/local/lib/summon/

uninstall:
	rm -f /usr/local/lib/summon/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

release: clean
	mkdir -p dist
	$(foreach PLATFORM,$(PLATFORMS),\
		$(foreach ARCH,$(ARCHITECTURES),\
			GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o dist/$(BINARY_NAME)-$(PLATFORM)-$(ARCH) -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH) && \
			tar -czf dist/$(BINARY_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH).tar.gz -C dist $(BINARY_NAME)-$(PLATFORM)-$(ARCH) && \
			rm -f dist/$(BINARY_NAME)-$(PLATFORM)-$(ARCH); \
		) \
	)
	@echo "Release packages created in dist/"

# Add a symlink to make it easier to use during development
dev-link: build
	mkdir -p $(HOME)/.config/summon/providers
	ln -sf $(PWD)/$(BINARY_NAME) $(HOME)/.config/summon/providers/

help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install the binary to /usr/local/lib/summon/"
	@echo "  uninstall  - Remove the binary from /usr/local/lib/summon/"
	@echo "  test       - Run tests"
	@echo "  clean      - Remove binary and dist directory"
	@echo "  release    - Create release packages for all platforms"
	@echo "  dev-link   - Create a symlink for development use"
	@echo "  help       - Show this help message"