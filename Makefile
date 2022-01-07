mock-gen:
	go generate ./...

coverage:
	go test ./... -coverprofile cover.out

test:
	go test ./...

lint:
	golangci-lint run