.PHONY: all clean build

all: clean build

clean:
	go clean -i ./...

build:
	CGO_ENABLED=0 go build -ldflags '-extldflags "-static"'

maildirs:
	for i in {1..1000}; do \
		mkdir -p maildirs/user$$i@example.com/{cur,new,tmp}; \
	done
