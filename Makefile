BIN = dive
TEMP_DIR = ./.tmp
PWD := ${CURDIR}
REGISTRY ?= docker.io
SHELL = /bin/bash -o pipefail
TEST_IMAGE = busybox:latest

# Tool versions #################################
GOLANG_CI_VERSION = v1.61.0
GOBOUNCER_VERSION = v0.4.0
GORELEASER_VERSION = v2.4.4
GOSIMPORTS_VERSION = v0.3.8
CHRONICLE_VERSION = v0.8.0
GLOW_VERSION = v1.5.1
DOCKER_CLI_VERSION = 27.3.1

# Command templates #################################
LINT_CMD = $(TEMP_DIR)/golangci-lint run --tests=false --timeout=2m --config .golangci.yaml
GOIMPORTS_CMD = $(TEMP_DIR)/gosimports -local github.com/joschi
RELEASE_CMD = DOCKER_CLI_VERSION=$(DOCKER_CLI_VERSION) $(TEMP_DIR)/goreleaser release --clean
SNAPSHOT_CMD = $(RELEASE_CMD) --skip=publish --skip=sign --snapshot
CHRONICLE_CMD = $(TEMP_DIR)/chronicle
GLOW_CMD = $(TEMP_DIR)/glow

# Formatting variables #################################
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)

# Test variables #################################
# the quality gate lower threshold for unit test total % coverage (by function statements)
COVERAGE_THRESHOLD := 30

## Build variables #################################
DIST_DIR = dist
SNAPSHOT_DIR = snapshot
OS=$(shell uname | tr '[:upper:]' '[:lower:]')
SNAPSHOT_BIN=$(realpath $(shell pwd)/$(SNAPSHOT_DIR)/$(OS)-build_$(OS)_amd64_v1/$(BIN))
CHANGELOG := CHANGELOG.md
VERSION=$(shell git describe --dirty --always --tags)

ifeq "$(strip $(VERSION))" ""
 override VERSION = $(shell git describe --always --tags --dirty)
endif

## Variable assertions

ifndef TEMP_DIR
	$(error TEMP_DIR is not set)
endif

ifndef DIST_DIR
	$(error DIST_DIR is not set)
endif

ifndef SNAPSHOT_DIR
	$(error SNAPSHOT_DIR is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef


.PHONY: all
all: clean static-analysis test ## Run all static analysis and tests
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

.PHONY: test
test: unit ## Run all tests (currently unit and cli tests)

$(TEMP_DIR):
	mkdir -p $(TEMP_DIR)


## Bootstrapping targets #################################

.PHONY: bootstrap-tools
bootstrap-tools: $(TEMP_DIR)
	$(call title,Bootstrapping tools)
	curl -sSfL https://raw.githubusercontent.com/anchore/chronicle/main/install.sh | sh -s -- -b $(TEMP_DIR)/ $(CHRONICLE_VERSION)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TEMP_DIR)/ $(GOLANG_CI_VERSION)
	curl -sSfL https://raw.githubusercontent.com/wagoodman/go-bouncer/master/bouncer.sh | sh -s -- -b $(TEMP_DIR)/ $(GOBOUNCER_VERSION)
	GOBIN="$(realpath $(TEMP_DIR))" go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)
	GOBIN="$(realpath $(TEMP_DIR))" go install github.com/rinchsan/gosimports/cmd/gosimports@$(GOSIMPORTS_VERSION)
	GOBIN="$(realpath $(TEMP_DIR))" go install github.com/charmbracelet/glow@$(GLOW_VERSION)

.PHONY: bootstrap-go
bootstrap-go:
	$(call title,Bootstrapping go dependencies)
	go mod download

.PHONY: bootstrap
bootstrap: bootstrap-go bootstrap-tools ## Download and install all go dependencies (+ prep tooling in the ./tmp dir)


## Development targets ###################################

#run: build
#	$(BUILD_PATH) build -t dive-example:latest -f .data/Dockerfile.example .
#
#run-large: build
#	$(BUILD_PATH) amir20/clashleaders:latest
#
#run-podman: build
#	podman build -t dive-example:latest -f .data/Dockerfile.example .
#	$(BUILD_PATH) localhost/dive-example:latest --engine podman
#
#run-podman-large: build
#	$(BUILD_PATH) docker.io/amir20/clashleaders:latest --engine podman
#
#run-ci: build
#	CI=true $(BUILD_PATH) dive-example:latest --ci-config .data/.dive-ci
#
#dev:
#	docker run -ti --rm -v $(PWD):/app -w /app -v dive-pkg:/go/pkg/ golang:1.13 bash
#
#build: gofmt
#	go build -o $(BUILD_PATH)

.PHONY: generate-test-data
generate-test-data:
	docker build -t dive-test:latest -f .data/Dockerfile.test-image . && docker image save -o .data/test-docker-image.tar dive-test:latest && echo 'Exported test data!'


## Static analysis targets #################################

.PHONY: static-analysis
static-analysis: lint check-go-mod-tidy check-licenses

.PHONY: lint
lint: ## Run gofmt + golangci lint checks
	$(call title,Running linters)
	# ensure there are no go fmt differences
	@printf "files with gofmt issues: [$(shell gofmt -l -s .)]\n"
	@test -z "$(shell gofmt -l -s .)"

	# run all golangci-lint rules
	$(LINT_CMD)
	@[ -z "$(shell $(GOIMPORTS_CMD) -d .)" ] || (echo "goimports needs to be fixed" && false)

	# go tooling does not play well with certain filename characters, ensure the common cases don't result in future "go get" failures
	$(eval MALFORMED_FILENAMES := $(shell find . | grep -e ':'))
	@bash -c "[[ '$(MALFORMED_FILENAMES)' == '' ]] || (printf '\nfound unsupported filename characters:\n$(MALFORMED_FILENAMES)\n\n' && false)"

.PHONY: format
format: ## Auto-format all source code
	$(call title,Running formatters)
	gofmt -w -s .
	$(GOIMPORTS_CMD) -w .
	go mod tidy

.PHONY: lint-fix
lint-fix: format  ## Auto-format all source code + run golangci lint fixers
	$(call title,Running lint fixers)
	$(LINT_CMD) --fix

.PHONY: check-licenses
check-licenses:
	$(TEMP_DIR)/bouncer check ./...

check-go-mod-tidy:
	@ .github/scripts/go-mod-tidy-check.sh && echo "go.mod and go.sum are tidy!"


## Testing targets #################################

.PHONY: unit
unit: $(TEMP_DIR)  ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -race -coverprofile $(TEMP_DIR)/unit-coverage-details.txt ./...
	@.github/scripts/coverage.py $(COVERAGE_THRESHOLD) $(TEMP_DIR)/unit-coverage-details.txt


## Acceptance testing targets (CI only) #################################

# todo: add --pull=never when supported by host box
.PHONY: ci-test-docker-image
ci-test-docker-image:
	docker run \
		--rm \
		-t \
		-v /var/run/docker.sock:/var/run/docker.sock \
		'${REGISTRY}/joschi/dive:latest-amd64' \
			'${TEST_IMAGE}' \
			--ci

.PHONY: ci-test-deb-package-install
ci-test-deb-package-install:
	docker run \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /${PWD}:/src \
		-w /src \
		ubuntu:latest \
			/bin/bash -x -c "\
				apt update && \
				apt install -y curl && \
				curl -L 'https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_CLI_VERSION}.tgz' | \
					tar -vxzf - docker/docker --strip-component=1 && \
					mv docker /usr/local/bin/ &&\
				docker version && \
				apt install ./snapshot/dive_*_linux_amd64.deb -y && \
				dive --version && \
				dive '${TEST_IMAGE}' --ci \
			"

.PHONY: ci-test-deb-package-install
ci-test-rpm-package-install:
	docker run \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /${PWD}:/src \
		-w /src \
		fedora:latest \
			/bin/bash -x -c "\
				curl -L 'https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_CLI_VERSION}.tgz' | \
					tar -vxzf - docker/docker --strip-component=1 && \
					mv docker /usr/local/bin/ &&\
				docker version && \
				dnf install ./snapshot/dive_*_linux_amd64.rpm -y && \
				dive --version && \
				dive '${TEST_IMAGE}' --ci \
			"

.PHONY: ci-test-linux-run
ci-test-linux-run:
	ls -la $(SNAPSHOT_DIR)
	ls -la $(SNAPSHOT_DIR)/dive_linux_amd64_v1
	chmod 755 $(SNAPSHOT_DIR)/dive_linux_amd64_v1/dive && \
	$(SNAPSHOT_DIR)/dive_linux_amd64_v1/dive '${TEST_IMAGE}'  --ci && \
    $(SNAPSHOT_DIR)/dive_linux_amd64_v1/dive --source docker-archive .data/test-kaniko-image.tar  --ci --ci-config .data/.dive-ci

# we're not attempting to test docker, just our ability to run on these systems. This avoids setting up docker in CI.
.PHONY: ci-test-mac-run
ci-test-mac-run:
	chmod 755 $(SNAPSHOT_DIR)/dive_darwin_amd64_v1/dive && \
	$(SNAPSHOT_DIR)/dive_darwin_amd64_v1/dive --source docker-archive .data/test-docker-image.tar  --ci --ci-config .data/.dive-ci

# we're not attempting to test docker, just our ability to run on these systems. This avoids setting up docker in CI.
.PHONY: ci-test-windows-run
ci-test-windows-run:
	dive.exe --source docker-archive .data/test-docker-image.tar  --ci --ci-config .data/.dive-ci


## Build-related targets #################################

.PHONY: build
build: $(SNAPSHOT_DIR) ## Build release snapshot binaries and packages

$(SNAPSHOT_DIR): ## Build snapshot release binaries and packages
	$(call title,Building snapshot artifacts)

	@# create a config with the dist dir overridden
	@echo "dist: $(SNAPSHOT_DIR)" > $(TEMP_DIR)/goreleaser.yaml
	@cat .goreleaser.yaml >> $(TEMP_DIR)/goreleaser.yaml

	@# build release snapshots
	@bash -c "\
		VERSION=$(VERSION:v%=%) \
		$(SNAPSHOT_CMD) --config $(TEMP_DIR)/goreleaser.yaml \
	  "

.PHONY: cli
cli: $(SNAPSHOT_DIR) ## Run CLI tests
	chmod 755 "$(SNAPSHOT_BIN)"
	$(SNAPSHOT_BIN) version
	go test -count=1 -timeout=15m -v ./test/cli

.PHONY: changelog
changelog: clean-changelog  ## Generate and show the changelog for the current unreleased version
	$(CHRONICLE_CMD) -vvv -n --version-file VERSION > $(CHANGELOG)
	@$(GLOW_CMD) $(CHANGELOG)

$(CHANGELOG):
	$(CHRONICLE_CMD) -vvv > $(CHANGELOG)

.PHONY: release
release:  ## Cut a new release
	@.github/scripts/trigger-release.sh

.PHONY: ci-release
ci-release: ci-check clean-dist $(CHANGELOG)
	$(call title,Publishing release artifacts)

	# create a config with the dist dir overridden
	echo "dist: $(DIST_DIR)" > $(TEMP_DIR)/goreleaser.yaml
	cat .goreleaser.yaml >> $(TEMP_DIR)/goreleaser.yaml

	bash -c "$(RELEASE_CMD) --release-notes <(cat CHANGELOG.md) --config $(TEMP_DIR)/goreleaser.yaml"

.PHONY: ci-check
ci-check:
	@.github/scripts/ci-check.sh


## Cleanup targets #################################

.PHONY: clean
clean: clean-dist clean-snapshot  ## Remove previous builds, result reports, and test cache

.PHONY: clean-snapshot
clean-snapshot:
	rm -rf $(SNAPSHOT_DIR) $(TEMP_DIR)/goreleaser.yaml

.PHONY: clean-dist
clean-dist: clean-changelog
	rm -rf $(DIST_DIR) $(TEMP_DIR)/goreleaser.yaml

.PHONY: clean-changelog
clean-changelog:
	rm -f $(CHANGELOG) VERSION


## Help! #################################

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'
