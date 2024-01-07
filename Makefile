version := $(shell git rev-parse HEAD)

cli:
	go build -ldflags "-X github.com/getsavvyinc/savvy-cli/config.version=$(version)" -o savvy .

cli_dev:
	go build -ldflags "-X github.com/getsavvyinc/savvy-cli/config.version=$(version)" -tags dev -o savvy-dev .

release:
	goreleaser release --clean

build_all:
	goreleaser build --clean
