BIN = die

all: clean build

run: build
	./build/$(BIN)

build: deps
	go build -o build/$(BIN) ./cmd/...

install: deps
	go install ./...

deps:
	command -v $(GOPATH)/bin/dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	$(GOPATH)/bin/dep ensure

test: build
	@! git grep tcell -- ':!tui/' ':!Gopkg.lock' ':!Gopkg.toml' ':!Makefile'
	go test -v ./...

lint: lintdeps build
	golint -set_exit_status $$(go list ./... | grep -v /vendor/)

lintdeps:
	go get -d -v -t ./...
	command -v golint >/dev/null || go get -u github.com/golang/lint/golint

clean:
	rm -rf build
	rm -rf vendor
	go clean

.PHONY: build install deps test lint lintdeps clean
