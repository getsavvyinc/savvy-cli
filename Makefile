cli:
	go build -o savvy .

cli_dev:
	go build -tags dev -o savvy-dev .

release:
	goreleaser release
