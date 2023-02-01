BINARY=bin/server

build:
	chmod +x initFileSystem.sh & ./initFileSystem.sh
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY} server.go
 	GOARCH=amd64 GOOS=linux go build -o ${BINARY} server.go
 	GOARCH=amd64 GOOS=windows go build -o ${BINARY} server.go

run: build
	./bin/server

test:
	go test -v

clean:
	go clean
	rm bin/*
	rm storage/*.txt
	./initFileSystem.sh