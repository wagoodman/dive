BIN = dive

all: clean build

run: build
	./build/$(BIN) build -t dive-example:latest -f .data/Dockerfile.example .

run-ci: build
	CI=true ./build/$(BIN) dive-example:latest --ci-config .data/.dive-ci

run-large: build
	./build/$(BIN) amir20/clashleaders:latest

build:
	go build -o build/$(BIN)

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
	grep -R 'const allowTestDataCapture = false' ui/
	go vet ./...
	@! gofmt -s -l . 2>&1 | grep -vE '^\.git/' | grep -vE '^\.cache/'
	golangci-lint run

lint: build
	golint -set_exit_status $$(go list ./...)

generate-test-data:
	docker build -t dive-test:latest -f .data/Dockerfile.test-image . && docker image save -o .data/test-docker-image.tar dive-test:latest && echo "Exported test data!"

clean:
	rm -rf build
	go clean

.PHONY: build install test lint clean release validate generate-test-data test-coverage ci
