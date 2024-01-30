
build:
	go build -o dist/dodo ./src/main.go

install-local:
	goreleaser release --snapshot --clean
	sudo cp ./dist/dodo-client_linux_amd64_v1/dodo-client /usr/local/bin