all: check lint

check:
	go test ./...

fmt:
	go fmt ./...

lint:
	golint ./...
	staticcheck ./..
