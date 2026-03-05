# Dive MCP Server: Roadmap to Production

This document outlines the strategy for addressing the identified gaps in the initial MCP implementation, categorized into three focused phases.

## Phase 1: Protocol Maturity & Resources
**Goal:** Align fully with the MCP 2025-11-25 standard and provide stable reference points for images.

*   **Implement Resource Registry:**
    *   Map `image.Analysis` to `dive://image/{name}/summary`.
    *   Expose `dive://image/{name}/efficiency` as a dynamic resource.
    *   **Rationale:** Allows agents to "look up" image state without re-triggering a tool call, enabling better context management in LLMs.
*   **Prompt Templates:**
    *   Implement `optimize-dockerfile`: A template that injects the results of `get_wasted_space` into a system prompt for the AI.
    *   Implement `explain-layer`: A template to help users understand complex `RUN` commands and their filesystem impact.

## Phase 2: Intelligence & Advanced Tooling
**Goal:** Transform the server from a summary provider to a structured data source with deep diffing capabilities.

*   **Structured Data Transition:**
    *   Introduce JSON-schema-compliant output for all tools.
    *   Enable agents to programmatically iterate over file lists, allowing for complex "chains of thought" regarding filesystem cleanup.
*   **The "Diff" Tool:**
    *   Implement `diff_layers(image, base_layer_index, target_layer_index)`.
    *   Leverage `filetree.CompareAndMark` to return a structured list of Additions, Modifications, and Deletions between any two layers.
    *   **Rationale:** This mirrors the most powerful feature of the Dive TUI, giving AI agents surgical precision in identifying where specific files were introduced or bloated.

## Phase 3: UX, Performance & Security
**Goal:** Ensure the server is robust, responsive, and safe for multi-user or remote environments.

*   **Progressive UX:**
    *   Bridge `internal/bus` (partybus) to MCP `notifications/progress`.
    *   Standardize progress tokens so clients like Claude Desktop can show an active loading state during image pulls.
*   **Bounded Caching:**
    *   Replace the simple map with an LRU (Least Recently Used) cache.
    *   Add a `--mcp-cache-ttl` flag to automatically evict stale analysis results.
*   **Security Sandboxing:**
    *   Add a `--sandbox` flag to restrict `docker-archive` lookups to a specific directory.
    *   Validate all image paths against the sandbox root to prevent directory traversal attacks in remote SSE modes.

## Implementation Priorities (Immediate Next Steps)
1.  **Update Structured Output:** Start returning JSON strings in `TextContent` to improve AI parsing.
2.  **Implement Progress Reporting:** Crucial for the "alive" feel of the integration.
3.  **Add Layer Diffing:** The missing link between the TUI and the MCP server.
