package mcp

import (
	"fmt"
	"net/http"

	"github.com/anchore/clio"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/internal/log"
)

func NewServer(id clio.Identification, opts options.MCP) *server.MCPServer {
	s := server.NewMCPServer(
		id.Name,
		id.Version,
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	h := newToolHandlers(opts)

	// --- Tools ---

	// 1. analyze_image tool
	analyzeTool := mcp.NewTool("analyze_image",
		mcp.WithDescription("Analyze a docker image and return efficiency metrics and layer details (JSON)"),
		mcp.WithString("image",
			mcp.Required(),
			mcp.Description("The name of the image to analyze (e.g., 'ubuntu:latest')"),
		),
		mcp.WithString("source",
			mcp.Description("The container engine to fetch the image from (docker, podman, docker-archive). Defaults to 'docker'."),
		),
	)
	s.AddTool(analyzeTool, h.analyzeImageHandler)

	// 2. get_wasted_space tool
	wastedSpaceTool := mcp.NewTool("get_wasted_space",
		mcp.WithDescription("Get the list of inefficient files that contribute to wasted space in the image (JSON)"),
		mcp.WithString("image",
			mcp.Required(),
			mcp.Description("The name of the image to get wasted space for"),
		),
		mcp.WithString("source",
			mcp.Description("The container engine to fetch the image from (docker, podman, docker-archive). Defaults to 'docker'."),
		),
	)
	s.AddTool(wastedSpaceTool, h.getWastedSpaceHandler)

	// 3. inspect_layer tool
	inspectLayerTool := mcp.NewTool("inspect_layer",
		mcp.WithDescription("Inspect the contents of a specific layer in an image (JSON)"),
		mcp.WithString("image",
			mcp.Required(),
			mcp.Description("The name of the image to inspect"),
		),
		mcp.WithNumber("layer_index",
			mcp.Required(),
			mcp.Description("The index of the layer to inspect (starting from 0)"),
		),
		mcp.WithString("source",
			mcp.Description("The container engine to fetch the image from (docker, podman, docker-archive). Defaults to 'docker'."),
		),
		mcp.WithString("path",
			mcp.Description("The path within the layer to inspect. Defaults to '/'."),
		),
	)
	s.AddTool(inspectLayerTool, h.inspectLayerHandler)

	// 4. diff_layers tool
	diffLayersTool := mcp.NewTool("diff_layers",
		mcp.WithDescription("Compare two layers in an image and return file changes (JSON)"),
		mcp.WithString("image",
			mcp.Required(),
			mcp.Description("The name of the image"),
		),
		mcp.WithNumber("base_layer_index",
			mcp.Required(),
			mcp.Description("The index of the base layer for comparison"),
		),
		mcp.WithNumber("target_layer_index",
			mcp.Required(),
			mcp.Description("The index of the target layer to compare against the base"),
		),
		mcp.WithString("source",
			mcp.Description("The container engine to fetch the image from (docker, podman, docker-archive). Defaults to 'docker'."),
		),
	)
	s.AddTool(diffLayersTool, h.diffLayersHandler)

	// --- Resources ---

	// 1. Summary resource template
	summaryTemplate := mcp.NewResourceTemplate("dive://image/{name}/summary", "Image Summary",
		mcp.WithTemplateDescription("Get a JSON summary of the image analysis"),
	)
	s.AddResourceTemplate(summaryTemplate, h.resourceSummaryHandler)

	// 2. Efficiency resource template
	efficiencyTemplate := mcp.NewResourceTemplate("dive://image/{name}/efficiency", "Image Efficiency",
		mcp.WithTemplateDescription("Get the efficiency score and wasted bytes for an image (JSON)"),
	)
	s.AddResourceTemplate(efficiencyTemplate, h.resourceEfficiencyHandler)

	// --- Prompts ---

	// 1. Optimize Dockerfile prompt
	s.AddPrompt(mcp.Prompt{
		Name:        "optimize-dockerfile",
		Description: "Get suggestions for optimizing a Dockerfile based on image analysis",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "image",
				Description: "The name of the image to optimize",
				Required:    true,
			},
		},
	}, h.promptOptimizeDockerfileHandler)

	// 2. Explain Layer prompt
	s.AddPrompt(mcp.Prompt{
		Name:        "explain-layer",
		Description: "Get an explanation of the impact of a specific image layer",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "image",
				Description: "The name of the image",
				Required:    true,
			},
			{
				Name:        "layer_index",
				Description: "The index of the layer to explain",
				Required:    true,
			},
		},
	}, h.promptExplainLayerHandler)

	return s
}

func Run(id clio.Identification, s *server.MCPServer, opts options.MCP) error {
	if opts.Transport == "stdio" {
		log.Infof("Starting MCP Stdio server")
		return server.ServeStdio(s)
	}

	host := opts.Host
	if host == "" {
		host = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", host, opts.Port)

	baseURLHost := opts.Host
	if baseURLHost == "" || baseURLHost == "0.0.0.0" {
		baseURLHost = "localhost"
	}
	baseURL := fmt.Sprintf("http://%s:%d", baseURLHost, opts.Port)

	if opts.Host == "0.0.0.0" {
		log.Warnf("Listening on 0.0.0.0 but baseURL is set to localhost. Remote MCP clients might fail to connect. Consider setting --host to your actual IP or hostname.")
	}

	mux := http.NewServeMux()

	// Session extractor middleware to handle header normalization.
	// StreamableHTTPServer handles its own session logic, but we provide this
	// to ensure X-Mcp-Session-Id and other variants are normalized to Mcp-Session-Id.
	sessionMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Identify Session
			sessionID := r.Header.Get("Mcp-Session-Id")
			if sessionID == "" {
				sessionID = r.Header.Get("X-Mcp-Session-Id")
			}
			if sessionID == "" {
				sessionID = r.URL.Query().Get("sessionId")
			}

			// 2. Normalize Headers
			if sessionID != "" {
				// Ensure the standard header is set for downstream handlers
				r.Header.Set("Mcp-Session-Id", sessionID)
				w.Header().Set("Mcp-Session-Id", sessionID)
			}

			// 3. Handle Protocol Version
			if version := r.Header.Get("Mcp-Protocol-Version"); version != "" {
				w.Header().Set("Mcp-Protocol-Version", version)
			}

			next.ServeHTTP(w, r)
		})
	}

	switch opts.Transport {
	case "streamable-http", "sse":
		// Both transport options now use the modern Streamable HTTP implementation.
		// "sse" is maintained for backwards compatibility with setup scripts.
		endpoint := "/mcp"
		if opts.Transport == "sse" {
			endpoint = "/sse"
		}
		
		shs := server.NewStreamableHTTPServer(s, server.WithEndpointPath(endpoint))
		mux.Handle(endpoint, shs)
		
		// If transport is sse, also provide /message alias
		if opts.Transport == "sse" {
			mux.Handle("/message", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = "/sse"
				shs.ServeHTTP(w, r)
			}))
		}
		
		// Also support root and /mcp as aliases for convenience
		if endpoint != "/" {
			mux.Handle("/", shs)
		}
		if endpoint != "/mcp" {
			mux.Handle("/mcp", shs)
		}

	default:
		return fmt.Errorf("unsupported transport: %s", opts.Transport)
	}

	// Add CORS and global logging middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("MCP Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id, X-Mcp-Session-Id, Mcp-Protocol-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id, Mcp-Protocol-Version")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		sessionMiddleware(mux).ServeHTTP(w, r)
	})

	log.Infof("Starting MCP %s server on %s", opts.Transport, addr)
	fmt.Printf("Starting MCP %s server on %s\n", opts.Transport, addr)
	if opts.Transport == "streamable-http" {
		fmt.Printf("- Endpoint: %s/mcp\n", baseURL)
	} else {
		fmt.Printf("- SSE endpoint: %s/sse\n", baseURL)
		fmt.Printf("- Message endpoint: %s/message\n", baseURL)
	}

	return http.ListenAndServe(addr, handler)
}

