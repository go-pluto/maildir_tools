.PHONY: all clean build

all: clean build

clean:
	go clean -i ./...

build:
	CGO_ENABLED=0 go build -ldflags '-extldflags "-static"'
