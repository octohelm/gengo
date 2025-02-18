test:
	go test -v ./pkg/...

cover:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./pkg/...

fmt:
	go tool gofumpt -l -w .

dep:
	go get -u ./...