.PHONY: build test run clean

build:
	go build -o bin/courseviewer main.go

test:
	go test -v ./...

run:
	go run main.go

clean:
	rm -rf bin/