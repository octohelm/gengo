test:
	CGO_ENABLED=0 go test -failfast -v --count=1 ./...

test-race:
	CGO_ENABLED=1 go test -v -race --count=1 ./...

fmt:
	go tool gofumpt -l -w .

dep:
	go mod tidy

update:
    go get -u ./...
