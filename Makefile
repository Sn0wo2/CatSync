BINARY_NAME=CatSync
TAGS=catsync_all
LDFLAGS=-s -w
GOFLAGS=-trimpath

.PHONY: build run clean release schema test_goreleaser

build:
	@printf 'Building CatSync (all features)...\n'
	@go build $(GOFLAGS) -tags $(TAGS) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) ./cmd

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

release:
	@echo Running release tool...
	@python ./scripts/release/tag.py

schema:
	@echo Generating schema...
	@go run ./scripts/schema/main.go

test_goreleaser:
	@echo Running goreleaser...
	@goreleaser release --snapshot --clean

-include Makefile.local
