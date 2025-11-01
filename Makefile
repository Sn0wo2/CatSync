.PHONY: release
.PHONY: test_goreleaser

release:
	@echo Running release tool...
	@go run ./tool/release/main.go

test_goreleaser:
	@echo Running goreleaser...
	@goreleaser release --snapshot --clean
