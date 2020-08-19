default:
	echo 'Hello, world!'
test:
        go test -cover -race -count=1 ./... -v
lint:
        golangci-lint run ./...
