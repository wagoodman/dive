BIN = dive
BUILD_DIR = ./dist/dive_linux_amd64
BUILD_PATH = $(BUILD_DIR)/$(BIN)
PWD := ${CURDIR}

all: clean build

run: build
	$(BUILD_PATH) build -t dive-example:latest -f .data/Dockerfile.example .

run-ci: build
	CI=true $(BUILD_PATH) dive-example:latest --ci-config .data/.dive-ci

run-large: build
	$(BUILD_PATH) amir20/clashleaders:latest

build:
	go build -o $(BUILD_PATH)

release: test-coverage validate
	./.scripts/tag.sh
	goreleaser --rm-dist

install:
	go install ./...

ci: clean validate test-coverage

test: build
	go test -cover -v -race ./...

test-coverage: build
	./.scripts/test-coverage.sh

validate:
	grep -R 'const allowTestDataCapture = false' runtime/ui/
	go vet ./...
	@! gofmt -s -l . 2>&1 | grep -vE '^\.git/' | grep -vE '^\.cache/'
	golangci-lint run

lint: build
	golint -set_exit_status $$(go list ./...)

generate-test-data:
	docker build -t dive-test:latest -f .data/Dockerfile.test-image . && docker image save -o .data/test-docker-image.tar dive-test:latest && echo "Exported test data!"

setup:
	go get ./...
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b /go/bin v1.18.0

dev:
	docker run -ti --rm -v $(PWD):/app -w /app -v dive-pkg:/go/pkg/ golang:1.13 bash

clean:
	rm -rf dist
	go clean

.PHONY: build install test lint clean release validate generate-test-data test-coverage ci dev
