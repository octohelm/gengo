test:
	CGO_ENABLED=0 go test -failfast -v ./...

test.race:
	CGO_ENABLED=1 go test -v -race ./...

fmt:
	go tool gofumpt -l -w .

dep:
	go get -u ./...