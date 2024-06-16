
.PHONY: build
build:
	go build -o dist/dodo ./src/main.go

.PHONY: test
test:
	go test ./src

.PHONY: install-local
install-local:
	goreleaser release --snapshot --clean
	sudo cp ./dist/dodo-client_linux_amd64_v1/dodo-client /usr/local/bin

.PHONY: release
release:
	goreleaser release

.PHONY: fmt
fmt:
	golangci-lint run --fix src

.PHONY: clean
clean:
	rm -rf dist