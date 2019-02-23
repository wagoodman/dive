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

release: test validate
	./.scripts/tag.sh
	goreleaser --rm-dist

install:
	go install ./...

test: build
	go test -cover -v ./...

coverage: build
	./.scripts/test.sh

validate:
	grep -R 'const allowTestDataCapture = false' ui/
	go vet ./...
	@! gofmt -s -d -l . 2>&1 | grep -vE '^\.git/'

lint: build
	golint -set_exit_status $$(go list ./...)

generate-test-data:
	docker build -t dive-test:latest -f .data/Dockerfile.test-image . && docker image save -o .data/test-docker-image.tar dive-test:latest && echo "Exported test data!"

clean:
	rm -rf build
	rm -rf vendor
	go clean

.PHONY: build install test lint clean release validate generate-test-data
