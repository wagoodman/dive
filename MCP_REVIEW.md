# Dive MCP Server: Architecture and Code Review

## 1. Overview
This document provides a deep architectural and code review of the Model Context Protocol (MCP) server implementation in the `dive` project.

## 2. Git Changes and Design Mapping

| Phase | Design Task | Implementation Path |
| :--- | :--- | :--- |
| **Phase 1** | Scaffolding & Dependencies | `go.mod`, `cmd/dive/cli/cli.go`, `cmd/dive/cli/internal/options/mcp.go`, `cmd/dive/cli/internal/options/application.go` |
| **Phase 2** | Server & Transport | `cmd/dive/cli/internal/mcp/server.go`, `cmd/dive/cli/internal/command/mcp.go` |
| **Phase 3** | Tool Handlers & Data | `cmd/dive/cli/internal/mcp/handlers.go` |
| **Phase 4** | Validation | `cmd/dive/cli/internal/mcp/handlers_test.go` |
| **Phase 5** | Documentation | `README.md` |

## 3. Architectural Review

### 3.1 Alignment with Dive Design
The implementation strictly follows the "Headless Consumer" pattern. By treating the MCP server as a standalone command that reuses existing adapters (`Analyzer`, `Resolver`), the core domain remains untouched, ensuring zero impact on the TUI or CI logic.

### 3.2 Decoupling and Event Bus
*   **Success:** The MCP logic is isolated in `cmd/dive/cli/internal/mcp`.
*   **Observation:** The implementation currently bypasses the `partybus` for the final tool responses to ensure synchronous JSON-RPC compliance. However, it still triggers analysis logs via the standard `internal/log` which is consistent with the CLI's behavior.

### 3.3 Concurrency and State
*   The use of `sync.RWMutex` in `toolHandlers` correctly handles concurrent tool calls from MCP clients.
*   The session-level cache in `analyses map[string]*image.Analysis` prevents expensive re-analysis of the same image within a single session.

## 4. Code Review

### 4.1 Coding Standards
*   **Go 1.24:** Implementation uses modern Go patterns.
*   **Clio Integration:** The `MCP` command is correctly integrated using `app.SetupCommand`, and options are integrated into the main `Application` struct.
*   **Error Handling:** Proper use of `mcp.NewToolResultError` ensures that errors are communicated back to the LLM in a protocol-compliant manner.

### 4.2 Implementation Highlights
*   **Auto-Analysis:** The `getAnalysis` helper simplifies tool handlers by ensuring the image is analyzed if it hasn't been already. This improves the UX for AI agents that might call `get_wasted_space` before `analyze_image`.
*   **Transport Flexibility:** Support for both `stdio` and `sse` via a CLI flag allows for both local and remote usage.

### 4.3 Areas for Improvement
*   **Test Alignment:** `TestHandlers_GetWastedSpace_NoCache` needs to be updated because the handler now performs auto-analysis instead of failing.
*   **Data Pruning:** While `get_wasted_space` limits results to 20 files, `inspect_layer` also limits to 100 entries. This is good, but for extremely deep trees, a more sophisticated pagination might be needed in the future.

## 5. Test Review

### 5.1 Test Fulfillments
The unit tests cover the primary logic of tool handlers and cache interaction.

### 5.2 Identified Regression
*   **`TestHandlers_GetWastedSpace_NoCache`:** Failed during review because the implementation became "smarter" than the test (auto-analysis). This is a positive regression in functionality but requires a test update.

## 6. Identified Gaps and Future Work

To reach a production-ready state and fully leverage the MCP protocol, the following gaps should be addressed:

### 6.1 Protocol Feature Completeness
*   **Resources:** Implement the `dive://` URI scheme (e.g., `dive://image/{name}/summary`) to allow agents to reference image states as static or dynamic documents.
*   **Prompts:** Add pre-defined prompt templates like `optimize-dockerfile` to guide AI agents in interpreting analysis results.

### 6.2 Data Granularity
*   **Structured Output:** Transition from pure `TextContent` to structured JSON results. This would allow agents to programmatically process file lists and metadata rather than relying on regex parsing of text summaries.

### 6.3 Progressive UX
*   **Progress Notifications:** Bridge `internal/bus` (partybus) events to MCP `notifications/progress`. This is critical for large images where analysis can take significant time, providing feedback to the AI and user.

### 6.4 Advanced Analysis Tools
*   **Layer Diffing:** Add a `diff_layers(image, layer_a, layer_b)` tool to specifically highlight what changed between two points in the image history, mirroring `dive`'s core TUI capability.

### 6.5 System Stability and Security
*   **Cache Management:** The current in-memory cache is unbounded. Implement an LRU (Least Recently Used) cache or TTL-based eviction to prevent memory exhaustion in long-running SSE sessions.
*   **Security Sandboxing:** Implement path restriction for the `docker-archive` source to ensure the server cannot be used to read arbitrary files outside of a designated workspace.

## 7. Conclusion
The implementation is of high quality, respects all architectural boundaries of the `dive` project, and provides a robust foundation for AI-assisted container optimization. Addressing the identified gaps will transform it from a utility into a comprehensive knowledge provider for the AI ecosystem.
