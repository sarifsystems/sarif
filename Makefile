export GO111MODULE=on
BINARY_NAME=sarifd

all: deps build

install:
	go install cmd/sarifd

build:
	go build cmd/sarifd

test:
	go test -v ./...

clean:
	go clean
	rm -f ${BINARY_NAME}

deps:
	go build -v ./...

upgrade:
	go get -u
