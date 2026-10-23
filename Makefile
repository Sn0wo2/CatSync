BINARY_NAME ?= $(notdir $(CURDIR))
TAGS ?= catsync_all

.PHONY: lint lint-fix test build run clean schema security ci release dry-release

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

test:
	go test -race -tags "$(TAGS)" ./...

build:
	go build -trimpath -tags "$(TAGS)" -ldflags "-s -w" -o $(BINARY_NAME) ./cmd

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

schema:
	go run ./scripts/schema/main.go

security:
	go run golang.org/x/vuln/cmd/govulncheck@latest -tags "$(TAGS)" ./...

ci: lint test security

release:
	goreleaser release --clean

dry-release:
	goreleaser release --snapshot --clean
