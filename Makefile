.PHONY: release
.PHONY: test_goreleaser
.PHONY: run_docker

release:
	@echo Running release tool...
	@go run ./tool/release/main.go

test_goreleaser:
	@echo Running goreleaser...
	@goreleaser release --snapshot --clean

run_docker:
	@echo Running docker compose...
	@docker-compose -f docker/docker-compose.yml up -d
