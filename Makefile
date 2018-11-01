BIN = dive

all: clean build

run: build
	./build/$(BIN) build -t dive-test:latest -f .data/Dockerfile .

run-large: build
	./build/$(BIN) amir20/clashleaders:latest

build:
	go build -o build/$(BIN)

release: test
	./.scripts/tag.sh
	goreleaser --rm-dist

install:
	go install ./...

test: build
	go test -cover -v ./...

lint: build
	golint -set_exit_status $$(go list ./...)

clean:
	rm -rf build
	rm -rf vendor
	go clean

.PHONY: build install test lint clean release
