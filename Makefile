.PHONY: test lint cover bench fmt vet ci

GO ?= go
PKG := ./...

fmt:
	gofmt -s -w .
	$(GO) mod tidy

vet:
	$(GO) vet $(PKG)

lint:
	golangci-lint run

test:
	$(GO) test -race -count=1 $(PKG)

cover:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic $(PKG)
	$(GO) tool cover -func=coverage.out

bench:
	$(GO) test -bench=. -benchmem -run=^$$ $(PKG)

ci: fmt vet lint test cover