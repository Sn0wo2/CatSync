.PHONY: release
.PHONY: test_goreleaser
.PHONY: run_docker
.PHONY: build
.PHONY: schema

release:
	@echo Running release tool...
	@python ./scripts/release/tag.py

schema:
	@echo Generating schema...
	@go run ./scripts/schema/main.go

build:
	@echo Building CatSync (all features)...
	@go build -trimpath -tags catsync_all -ldflags "-s -w" -o CatSync ./cmd

test_goreleaser:
	@echo Running goreleaser...
	@goreleaser release --snapshot --clean

run_docker:
	@echo Running docker compose...
	@docker-compose -f docker/docker-compose.yml up -d
