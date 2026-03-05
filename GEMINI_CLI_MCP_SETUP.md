# Guide: Connecting Dive MCP to Gemini-CLI

This guide explains how to configure Gemini-CLI to use the `dive` MCP server, enabling deep container image analysis directly within your chat sessions.

## 1. Start the Dive MCP Server

The recommended transport for modern MCP clients (like Gemini-CLI, Cursor, and Claude Desktop) is **Streamable HTTP**. This transport is more robust, handles sessions automatically, and is fully compliant with the latest MCP specification.

### Basic Startup (Streamable HTTP)
```bash
# Start the server on the default port (8080)
./dive mcp --transport streamable-http
```

### Alternative: SSE Startup
If you specifically need the legacy SSE transport:
```bash
./dive mcp --transport sse --port 8080
```

### Recommended Production Startup
Use the following command to enable security sandboxing and suppress non-protocol logs on stdout:
```bash
./dive mcp --transport streamable-http --port 8080 --mcp-sandbox $(pwd) --quiet
```

---

## 2. Configure Gemini-CLI

Gemini-CLI reads its MCP server configurations from its global configuration file.

### Locate your config
Usually found at: `~/.gemini-cli/config.yaml` (Linux/macOS) or `%USERPROFILE%\.gemini-cli\config.yaml` (Windows).

### Add the Dive Server
Add the following entry under the `mcpServers` key. 

**For Streamable HTTP (Recommended):**
```yaml
# ~/.gemini-cli/config.yaml

mcpServers:
  dive:
    url: "http://localhost:8080/mcp"
```

**For SSE (Legacy):**
```yaml
mcpServers:
  dive:
    url: "http://localhost:8080/sse"
```

*Note: If you are using the **Stdio** transport instead of HTTP, use this configuration:*
```yaml
mcpServers:
  dive:
    command: "/absolute/path/to/dive"
    args: ["mcp", "--quiet"]
```

---

## 3. Verify the Connection

Restart your Gemini-CLI session. Once started, verify that the tools are registered by asking the agent:

> **User:** "What MCP tools are currently available?"
>
> **Agent:** "I have access to the following tools from the **dive** server:
> - `analyze_image`: Analyze a docker image and return efficiency metrics.
> - `get_wasted_space`: Get the list of inefficient files.
> - `inspect_layer`: Inspect the contents of a specific layer.
> - `diff_layers`: Compare two layers and return file changes."

---

## 4. Troubleshooting: "Missing sessionId"

If you encounter a `Missing sessionId` error when using the SSE transport, it's likely because your client is attempting to send messages before establishing a session or is not correctly handling the MCP-specific SSE handshake.

**Solution:** Switch to the `streamable-http` transport (as shown in section 1 and 2), which is designed to handle these scenarios gracefully.

---

## 5. Example Usage in Gemini-CLI
... (rest of the file remains the same)

Once connected, you can use natural language to trigger deep analysis:

**Analyze a local image:**
> "Analyze the image 'my-app:latest' and tell me the efficiency score."

**Identify bloated files:**
> "Show me the top 10 most inefficient files in 'my-app:latest'."

**Compare build stages:**
> "Show me exactly what changed between layer 2 and layer 3 of my image."

**Optimize via Prompt:**
> "Use the 'optimize-dockerfile' prompt for image 'my-app:latest' and give me suggestions."

---

## 5. Persistent Server Settings (Optional)

To avoid typing flags every time you start the server, you can save your preferences in `~/.dive.yaml`:

```yaml
# ~/.dive.yaml
mcp:
  transport: sse
  port: 8080
  mcp-sandbox: /home/user/images
  mcp-cache-size: 20
  mcp-cache-ttl: 24h
```

Now, you can simply run:
```bash
dive mcp
```
And it will start as an SSE server with your predefined settings.
