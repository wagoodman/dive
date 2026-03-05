# Dive Project Architecture & Development Guide

## Overview
`dive` is a tool for exploring a docker image, layer contents, and discovering ways to shrink the size of your Docker/OCI image. It analyzes each layer of an image, showing the changes in the file system, and calculates an "efficiency" score based on wasted space (duplicated or deleted files across layers).

## Architecture

### 1. High-Level Components
The project follows a decoupled, event-driven architecture to separate the CLI logic, image analysis, and the Terminal User Interface (TUI).

*   **CLI (cmd/dive/cli):** Managed by `clio` and `cobra`. It handles configuration, flag parsing, and orchestrates the analysis flow.
*   **Image Domain (dive/image):** Core models for `Image`, `Layer`, and `Analysis`. It includes resolvers for different sources (Docker, Podman, Archives).
*   **Filetree Domain (dive/filetree):** The "brain" of the application. It manages the representation of file systems, handles "stacking" layers, and calculates efficiency.
*   **Internal Bus (internal/bus):** Powered by `go-partybus`. It allows the CLI and Analysis components to communicate with the UI via events (e.g., `ExploreAnalysis`, `TaskStarted`).
*   **TUI (cmd/dive/cli/internal/ui):** Built with `gocui` (and `lipgloss` for styling). It consumes events from the bus to render the interactive explorer.

### 2. Analysis Flow
1.  **Resolve:** The `GetImageResolver` determines the source (Docker Engine, Podman, or Tarball).
2.  **Fetch:** The resolver extracts the image and builds a list of `Layer` objects, each containing a `FileTree`.
3.  **Analyze:** The `Analyze` function (in `dive/image/analysis.go`) triggers the `Efficiency` calculation.
4.  **Efficiency Calculation:** 
    *   Iterates through all layer trees.
    *   Tracks file paths across layers.
    *   Identifies "wasted space" (files rewritten in subsequent layers or deleted files that still occupy space in lower layers).
5.  **Explore:** If not in CI mode, the `Analysis` results are published to the bus, triggering the TUI.

### 3. TUI Structure (V1)
The TUI uses an MVC-like pattern:
*   **App/Controller:** Manages the `gocui` main loop and coordinate views.
*   **Views:** Specialized components for `Layer`, `FileTree`, `ImageDetails`, etc.
*   **ViewModel:** Buffers and formats data for presentation.

## Engineering Standards

### Coding Standards
*   **Go Version:** 1.24.
*   **Linting:** Enforced via `golangci-lint` with a specific configuration in `.golangci.yaml`. 
*   **Formatting:** Standard `gofmt -s` and `go mod tidy`.
*   **CLI Framework:** Uses `github.com/anchore/clio` for standardized application setup (logging, configuration, versioning).

### Testing Infrastructure
*   **Unit Tests:** Standard `go test`. Coverage is tracked and enforced (threshold: 25%) via `.github/scripts/coverage.py`.
*   **CLI/Integration Tests:** Located in `cmd/dive/cli/`, using `go-snaps` for snapshot testing of CLI output and configuration.
*   **Acceptance Tests:** Automated cross-platform tests (Linux, Mac, Windows) that run the built binary against real test images (located in `.data/`).

### Quality Gates
*   **Static Analysis:** Gofmt check, file name validation (no `:`), and `golangci-lint`.
*   **License Compliance:** Checked via `bouncer`.
*   **CI Pipeline:** GitHub Actions (`validations.yaml`) runs on every PR, executing:
    *   Static Analysis.
    *   Unit Tests (with coverage check).
    *   Snapshot builds.
    *   Acceptance tests on multiple platforms.

## Build and Release
*   **Taskfile.yaml:** The primary entry point for development tasks (`task test`, `task build`, `task lint`).
*   **Makefile:** A shim for `Taskfile` for users accustomed to `make`.
*   **Goreleaser:** Manages the build matrix (Linux, Darwin, Windows across various archs), generates `.deb`, `.rpm`, Homebrew formulas, and Docker images.
*   **CI Release:** Automated via `.github/workflows/release.yaml` on tag pushes.

## Key Dependencies
*   `github.com/awesome-gocui/gocui`: TUI framework.
*   `github.com/wagoodman/go-partybus`: Event bus.
*   `github.com/anchore/clio`: Application framework.
*   `github.com/docker/docker`: Docker API integration.
*   `github.com/gkampitakis/go-snaps`: Snapshot testing.

## Future Development Notes
*   **Windows Support:** While present, acceptance tests for Windows are noted as "todo" in some areas or require specific runners.
*   **Performance:** The filetree stacking and efficiency calculation are CPU and memory intensive for very large images; optimizations should focus on `dive/filetree`.
*   **UI Modernization:** Styling is increasingly moving towards `lipgloss`, though the core remains `gocui`.
