BIN = dive
BUILD_DIR = ./dist/dive_linux_amd64
BUILD_PATH = $(BUILD_DIR)/$(BIN)
PWD := ${CURDIR}
PRODUCTION_REGISTRY = docker.io
TEST_IMAGE = busybox:latest

all: clean build

## For CI

ci-unit-test:
	go test -cover -v -race ./...

ci-static-analyses:
	grep -R 'const allowTestDataCapture = false' runtime/ui/viewmodel
	go vet ./...
	@! gofmt -s -l . 2>&1 | grep -vE '^\.git/' | grep -vE '^\.cache/'
	golangci-lint run

ci-install-go-tools:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sudo sh -s -- -b /usr/local/bin/ latest

ci-docker-login:
	echo '${DOCKER_PASSWORD}' | docker login -u '${DOCKER_USERNAME}' --password-stdin '${PRODUCTION_REGISTRY}'

ci-docker-logout:
	docker logout '${PRODUCTION_REGISTRY}'

ci-publish-release:
	goreleaser --rm-dist

# todo: add --pull=never when supported by host box
ci-test-production-image:
	docker run \
		--rm \
		-t \
		-v //var/run/docker.sock://var/run/docker.sock \
		'${PRODUCTION_REGISTRY}/wagoodman/dive:latest' \
			'${TEST_IMAGE}' \
			--ci


## For development

run: build
	$(BUILD_PATH) build -t dive-example:latest -f .data/Dockerfile.example .

run-large: build
	$(BUILD_PATH) amir20/clashleaders:latest

run-podman: build
	podman build -t dive-example:latest -f .data/Dockerfile.example .
	$(BUILD_PATH) localhost/dive-example:latest --engine podman

run-podman-large: build
	$(BUILD_PATH) docker.io/amir20/clashleaders:latest --engine podman

run-ci: build
	CI=true $(BUILD_PATH) dive-example:latest --ci-config .data/.dive-ci

build:
	go build -o $(BUILD_PATH)

generate-test-data:
	docker build -t dive-test:latest -f .data/Dockerfile.test-image . && docker image save -o .data/test-docker-image.tar dive-test:latest && echo 'Exported test data!'

dev:
	docker run -ti --rm -v $(PWD):/app -w /app -v dive-pkg:/go/pkg/ golang:1.13 bash

clean:
	rm -rf dist
	go clean


