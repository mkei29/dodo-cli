
build:
	go build -o dist/dodo ./src/main.go
.PHONY: build

test:
	go test ./src
.PHONY: test

install-local:
	goreleaser release --snapshot --clean
	sudo cp ./dist/dodo-client_linux_amd64_v1/dodo-client /usr/local/bin
.PHONY: install-local

.PHONY: release
release:
	goreleaser release

fmt:
	golangci-lint run --fix src