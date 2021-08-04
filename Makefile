include .env_file

# Go parameters
GOCMD=GO111MODULE=on GOFLAGS=-mod=vendor go
DEPCMD=$(GOCMD) mod
GOBUILD=$(GOCMD) build -ldflags "-X main.Version=$(VERSION) -linkmode=external"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

build: deps
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOCMD) test -cover ./...

lint:
	@golangci-lint run

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(DEPCMD) vendor

archive: clean build
	tar -czf $(BINARY_NAME).tar.gz $(BINARY_NAME)
