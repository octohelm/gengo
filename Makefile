test:
	go test -v ./pkg/...

test.race:
	CGO_ENABLED=1 go test -v -race ./pkg/...

fmt:
	go tool gofumpt -l -w .

dep:
	go get -u ./...