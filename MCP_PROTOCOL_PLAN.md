# Analysis and Implementation Plan: Standard Compliant MCP Protocol

## 1. Protocol Analysis

The Model Context Protocol (MCP) relies on a strict stateful lifecycle and standard JSON-RPC 2.0 messaging. The current implementation's "Missing sessionId" error stems from a mismatch between client expectations and server-side session tracking, particularly during the handshake phase.

### Core Lifecycle Requirements
1.  **Initialization Phase**: 
    - Client sends `initialize` (Request).
    - Server responds with `InitializeResult` (Response) + Capabilities + Protocol Version.
    - **Crucial**: In Streamable HTTP/SSE, the server MUST provide the `Mcp-Session-Id` header in this response if it hasn't been established yet.
    - Client sends `notifications/initialized` (Notification) to signal it's ready.
2.  **Session Persistence**: 
    - For SSE, the `sessionId` is typically assigned during the `GET /sse` request and then used in all subsequent `POST` requests.
    - For Streamable HTTP, the `sessionId` is assigned during the first `POST /initialize` and used thereafter.

### Identification of Fragility
The current implementation is fragile because:
- It manually intercepts `POST /sse` to handle `initialize` but fails to register the session in the underlying `mcp-go` session map.
- It returns a mocked JSON-RPC response for `initialize` that doesn't trigger the proper internal state transitions in the server library.
- It treats `sessionId` as a mandatory parameter for the middleware even before the session is fully established.

---

## 2. Implementation Plan

### Step 1: Unified Session Middleware
Refactor the `sessionMiddleware` to be less restrictive and more spec-compliant:
- **Header Priority**: Treat `Mcp-Session-Id` as the primary source of truth.
- **Lazy Injection**: Inject the `sessionId` into the query string ONLY if it exists, but do not fail the request if it's missing IF the method is `initialize`.
- **Protocol Versioning**: Pass-through the `Mcp-Protocol-Version` header to ensure compatibility with modern clients.

### Step 2: Protocol-Compliant SSE Handshake
- **Route /sse properly**: Standardize on the library's `SSEHandler` for GET and `MessageHandler` for POST.
- **Path Rewriting**: Continue rewriting `POST /sse` to `/message` but ensure the session is already active or being established.
- **Session Registration**: Ensure that every session ID generated is properly mapped to a `ClientSession` in the MCPServer.

### Step 3: Support for Standard JSON-RPC Methods
Ensure the server explicitly supports the following methods via the library or custom handlers:
- `initialize` (Lifecycle)
- `notifications/initialized` (Lifecycle)
- `ping` (Utility)
- `tools/list`, `tools/call` (Core Features)
- `resources/list`, `resources/read` (Core Features)
- `prompts/list`, `prompts/get` (Core Features)

### Step 4: Streamable HTTP Native Support
Fully leverage `server.NewStreamableHTTPServer` which is built specifically for the 2025-03-26 spec. This will handle the "POST-first" initialization correctly without custom logic.

---

## 3. Implementation Schedule

1.  **Phase 1: Refactor Middleware** (Fixing the "Missing sessionId" error by allowing `initialize` to pass through).
2.  **Phase 2: Modernize SSE Routing** (Using standard library handlers for `/sse` and `/message`).
3.  **Phase 3: Validation** (Using `curl` and `mcp-inspector` to verify the handshake according to the JSON-RPC 2.0 schema).
