# MCP Transport Update & Session Fixes

This document outlines the architectural changes and improvements made to the Dive MCP server to support the latest Model Context Protocol (MCP) 2025-03-26 specification.

## 1. Implementation of Streamable HTTP Transport

The server now supports the **Streamable HTTP** transport, which is the modern standard for MCP communication over HTTP. It consolidates communication into a single endpoint and provides robust session management.

### Key Features:
- **Single Endpoint:** Exposes a unified endpoint at `/mcp` (and `/` as an alias) that handles `GET` (for the SSE event stream), `POST` (for JSON-RPC messages), and `DELETE` (for session termination).
- **Session-First Design:** Automatically manages sessions via the `Mcp-Session-Id` header, as required by the latest specification.
- **Improved Robustness:** Eliminates the need for clients to manually track and provide session IDs in query parameters for every message.

### Usage:
Start the server with the new transport:
```bash
./dive mcp --transport streamable-http
```

---

## 2. Resolution of "Missing sessionId" Error

We identified and fixed the root cause of the `Missing sessionId` error that occurred when using the legacy SSE transport with modern MCP clients.

### Fixes Applied:
- **Robust Session Extraction:** The server now checks for session IDs in three locations to maximize compatibility:
    1. `Mcp-Session-Id` header (Modern spec)
    2. `X-Mcp-Session-Id` header (Common client variant)
    3. `sessionId` query parameter (Legacy spec)
- **Automatic Header Injection:** If a session ID is found in a header but missing from the query parameter, the server automatically injects it into the request context before passing it to the internal `mcp-go` handlers.
- **Initialization Handling:** Added special logging and bypass logic for initialization requests that do not yet have an assigned session ID.

---

## 3. Compatibility & Security Enhancements

To ensure the Dive MCP server works seamlessly with Gemini-CLI, Cursor, and other web-based MCP inspectors:

- **Enhanced CORS:** Added support for `DELETE` methods and explicitly exposed MCP-specific headers:
    - `Mcp-Session-Id`
    - `X-Mcp-Session-Id`
    - `Mcp-Protocol-Version`
- **Flexible Routing:** The SSE transport now supports direct `POST` requests to the `/sse` endpoint, a common behavior among clients that ignore the `endpoint` event in the SSE stream.
- **Clearer Networking Warnings:** Added proactive warnings when the server is bound to `0.0.0.0` but `baseURL` is set to `localhost`, helping users diagnose connectivity issues in remote or containerized environments.

---

## 4. Documentation Updates

- **`GEMINI_CLI_MCP_SETUP.md`:** Updated to recommend **Streamable HTTP** as the primary connection method for Gemini-CLI users.
- **CLI Help:** Updated the `dive mcp` command-line help to include `streamable-http` as a valid transport option.
