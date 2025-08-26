
.PHONY: build
build:
	go build -o dist/dodo ./src/main.go

.PHONY: test
test:
	go test ./src

.PHONY: fmt
fmt:
	golangci-lint run --fix src
	uv run ruff format
	uv run ruff check --fix

.PHONY: clean
clean:
	rm -rf dist
