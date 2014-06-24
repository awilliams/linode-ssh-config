BIN=linode-ssh-config
VERSION=0.1.0
README=README.md
LICENSE=LICENSE
EXAMPLE_INI=linode-ssh-config.ini.example
RELEASE_DIR=release

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOCMD) get -d -v ./... 
GOFMT=gofmt -w
 
default: build

build:
	$(GODEP)
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o bin/linux-amd64/$(BIN)
	GOARCH=386 GOOS=linux $(GOBUILD) -o bin/linux-386/$(BIN)
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o bin/darwin-amd64/$(BIN)

package:
	rm -rf $(RELEASE_DIR)/$(BIN)
	mkdir $(RELEASE_DIR)/$(BIN)
	cp $(README) $(RELEASE_DIR)/$(BIN)
	cp $(LICENSE) $(RELEASE_DIR)/$(BIN)
	cp $(EXAMPLE_INI) $(RELEASE_DIR)/$(BIN)/linode-ssh-config.example.ini

	cp -f bin/linux-amd64/$(BIN) $(RELEASE_DIR)/$(BIN)/$(BIN)
	tar -czf $(RELEASE_DIR)/$(BIN)-linux-amd64-v$(VERSION).tar.gz -C $(RELEASE_DIR) $(BIN)

	cp -f bin/linux-386/$(BIN) $(RELEASE_DIR)/$(BIN)/$(BIN)
	tar -czf $(RELEASE_DIR)/$(BIN)-linux-386-v$(VERSION).tar.gz -C $(RELEASE_DIR) $(BIN)

	cp -f bin/darwin-amd64/$(BIN) $(RELEASE_DIR)/$(BIN)/$(BIN)
	tar -czf $(RELEASE_DIR)/$(BIN)-darwin-amd64-v$(VERSION).tar.gz -C $(RELEASE_DIR) $(BIN)

	rm -rf $(RELEASE_DIR)/$(BIN)

format:
	$(GOFMT) ./**/*.go

clean:
	$(GOCLEAN)

test:
	$(GODEP) && $(GOTEST) -v ./...
