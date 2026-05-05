.PHONY: build install test vet lint fmt verify clean help

GO ?= go
BIN := ./bin/ripcord
PKG := ./...

help:
	@echo "Ripcord Makefile targets:"
	@echo "  build    Build $(BIN)"
	@echo "  install  go install ."
	@echo "  test     go test -race"
	@echo "  vet      go vet"
	@echo "  lint     golangci-lint run"
	@echo "  fmt      gofmt + goimports"
	@echo "  verify   vet + test + lint (run before pushing)"
	@echo "  clean    Remove build artifacts"

build:
	$(GO) build -o $(BIN) .

install:
	$(GO) install .

test:
	$(GO) test $(PKG) -count=1 -race

vet:
	$(GO) vet $(PKG)

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

verify: vet test lint

clean:
	rm -rf ./bin ./dist coverage.out coverage.html
