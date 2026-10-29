BINARY_NAME ?= CatSync
TAGS ?= catsync_all

.PHONY: lint lint-fix test build run clean schema security ci release dry-release

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run

lint-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run --fix

test:
	go test $(if $(filter 1,$(shell go env CGO_ENABLED)),-race) -tags "$(TAGS)" ./...

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
