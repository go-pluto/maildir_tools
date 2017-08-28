.PHONY: all clean build

all: clean build

clean:
	go clean -i ./...

build:
	CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o maildir_dumper ./cmd/dumper
	CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o maildir_visualizer ./cmd/visualizer

install:
	CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static"' ./cmd/dumper
	CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static"' ./cmd/visualizer

maildirs:
	for i in {1..1000}; do \
		mkdir -p maildirs/user$$i@example.com/{cur,new,tmp}; \
	done
