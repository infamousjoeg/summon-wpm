.PHONY: build install test clean release

BINARY_NAME=summon-wpm
VERSION=$(shell grep "version = " cmd/summon-wpm/main.go | cut -d '"' -f2)
MAIN_PATH=./cmd/summon-wpm

build:
	go build -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)

build-universal-macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	lipo -create -output $(BINARY_NAME) $(BINARY_NAME)-darwin-amd64 $(BINARY_NAME)-darwin-arm64
	rm $(BINARY_NAME)-darwin-amd64 $(BINARY_NAME)-darwin-arm64

install: build
	mkdir -p /usr/local/lib/summon
	cp $(BINARY_NAME) /usr/local/lib/summon/

install-universal-macos: build-universal-macos
	mkdir -p /usr/local/lib/summon
	cp $(BINARY_NAME) /usr/local/lib/summon/

uninstall:
	rm -f /usr/local/lib/summon/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-darwin-amd64
	rm -f $(BINARY_NAME)-darwin-arm64
	rm -rf dist/

release: clean
	mkdir -p dist
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C dist $(BINARY_NAME)-linux-amd64
	
	GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C dist $(BINARY_NAME)-linux-arm64
	
	# macOS Universal
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	lipo -create -output dist/$(BINARY_NAME)-darwin-universal dist/$(BINARY_NAME)-darwin-amd64 dist/$(BINARY_NAME)-darwin-arm64
	tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz -C dist $(BINARY_NAME)-darwin-universal
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	(cd dist && zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe)
	
	GOOS=windows GOARCH=arm64 go build -o dist/$(BINARY_NAME)-windows-arm64.exe -ldflags "-X main.version=$(VERSION)" $(MAIN_PATH)
	(cd dist && zip $(BINARY_NAME)-$(VERSION)-windows-arm64.zip $(BINARY_NAME)-windows-arm64.exe)
	
	@echo "Release packages created in dist/"

# Add a symlink to make it easier to use during development
dev-link: build
	mkdir -p $(HOME)/.config/summon/providers
	ln -sf $(PWD)/$(BINARY_NAME) $(HOME)/.config/summon/providers/

help:
	@echo "Available targets:"
	@echo "  build                - Build the binary for current platform"
	@echo "  build-universal-macos - Build a universal macOS binary (Intel + Apple Silicon)"
	@echo "  install              - Install the binary to /usr/local/lib/summon/"
	@echo "  install-universal-macos - Build and install a universal macOS binary"
	@echo "  uninstall            - Remove the binary from /usr/local/lib/summon/"
	@echo "  test                 - Run tests"
	@echo "  clean                - Remove binary and dist directory"
	@echo "  release              - Create release packages for all platforms"
	@echo "  dev-link             - Create a symlink for development use"
	@echo "  help                 - Show this help message"