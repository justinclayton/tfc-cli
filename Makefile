BINARY    := tfc
MODULE    := github.com/hashicorp/ddr/apps/tfc
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint clean install snapshot

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/

snapshot:
	goreleaser release --snapshot --clean
