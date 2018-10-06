BIN = dive

all: clean build

# karuppiah7890/bashful
#	docker image ls | grep "dive-test" >/dev/null || docker build -t dive-test:latest .
#	./build/$(BIN) die-test
run: build
	./build/$(BIN) karuppiah7890/bashful

build:
	go build -o build/$(BIN)

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

.PHONY: build install test lint clean
