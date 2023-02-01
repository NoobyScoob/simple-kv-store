BINARY=bin/server

build:
	chmod +x initFileSystem.sh & ./initFileSystem.sh
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY}_darwin server.go
	GOARCH=amd64 GOOS=linux go build -o ${BINARY}_linux server.go

run: build
	./bin/server_darwin

test:
	go test -v

clean:
	go clean
	rm bin/*
	rm storage/*.txt
	./initFileSystem.sh