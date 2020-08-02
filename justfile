default:
	echo 'Hello, world!'
test:
        go test -cover -race -count=1 ./...
