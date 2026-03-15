# Contributing to dive

Thank you for your interest in contributing to dive! This document covers how to get the project building and running locally, and how to submit your changes.

## Prerequisites

- [Go](https://go.dev/dl/) 1.24 or later
- [Docker](https://docs.docker.com/get-docker/) (required for integration/CLI tests)
- `make` or `task` (see [Bootstrapping](#bootstrapping) below)

## Getting the source

```bash
# Fork the repo on GitHub first, then:
git clone https://github.com/<your-username>/dive.git
cd dive
git remote add upstream https://github.com/wagoodman/dive.git
```

## Bootstrapping

The project uses [task](https://taskfile.dev) as its task runner, managed by [binny](https://github.com/anchore/binny). All tools are pinned and installed locally under `.tool/` — nothing is installed system-wide.

Bootstrap everything in one step:

```bash
make tools
```

From this point you can use either `make <target>` or `.tool/task <target>` interchangeably.

## Building

```bash
make build
```

This produces a local snapshot binary via [goreleaser](https://goreleaser.com). The binary ends up under `./snapshot/`. goreleaser is installed automatically by `make tools` via binny — no global install needed.

For a quick dev build without goreleaser:

```bash
go build -o dive ./cmd/dive
```

## Running the tests

```bash
# Run all validations (static analysis + tests)
make

# Unit tests only
make unit

# CLI/integration tests only (requires Docker)
make cli
```

## Static analysis

```bash
# Run linting and license checks
make static-analysis

# Just lint
make lint

# Check go.mod is tidy
make check-go-mod-tidy
```

Before submitting a PR, make sure `make` (the default target) passes cleanly.

## Project layout

```
cmd/          Entry points (CLI wiring via cobra)
  dive/       Main package (binary entry point)
dive/         Core library: image analysis, file-tree diffing
  filetree/   File-tree data structures and operations
  image/      Image fetch/parsing for Docker and Podman
internal/     Internal utilities
  bus/        Event bus
  log/        Logging
  utils/      Shared utility helpers
go.mod        Module definition and minimum Go version
.goreleaser.yaml  Release build configuration
.golangci.yaml    Linter configuration
.github/      CI workflows
```

## Submitting changes

1. Fork the repository and create a branch off `main`.
2. Make your changes. Keep commits focused and the diff minimal.
3. Run `make` and confirm everything passes.
4. Open a pull request against `wagoodman/dive:main`. Fill in a brief description of what you changed and why, and link any relevant issues.

## Code style

- Follow standard Go conventions (`gofmt`, `goimports`).
- The project uses [golangci-lint](https://golangci-lint.run/) with the configuration in `.golangci.yaml`. Running `make lint` locally will catch most issues before CI does.
- Match the style of the surrounding code when in doubt.

## License

By contributing you agree that your changes will be licensed under the [MIT License](LICENSE).
