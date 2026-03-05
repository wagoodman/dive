package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/command/adapter"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/internal/log"
)

type cachedAnalysis struct {
	Analysis  *image.Analysis
	Timestamp time.Time
}

type toolHandlers struct {
	opts     options.MCP
	analyses *lru.Cache[string, cachedAnalysis]
}

func newToolHandlers(opts options.MCP) *toolHandlers {
	cacheSize := opts.CacheSize
	if cacheSize <= 0 {
		cacheSize = 10
	}
	cache, _ := lru.New[string, cachedAnalysis](cacheSize)
	return &toolHandlers{
		opts:     opts,
		analyses: cache,
	}
}

// --- Data Models for Structured Output ---

type ImageSummary struct {
	Image           string         `json:"image"`
	TotalSize       uint64         `json:"total_size_bytes"`
	EfficiencyScore float64        `json:"efficiency_score"`
	WastedSpace     uint64         `json:"wasted_space_bytes"`
	LayerCount      int            `json:"layer_count"`
	Layers          []LayerSummary `json:"layers"`
}

type LayerSummary struct {
	Index   int    `json:"index"`
	ID      string `json:"id"`
	Size    uint64 `json:"size_bytes"`
	Command string `json:"command"`
}

type WastedSpaceResult struct {
	Image           string             `json:"image"`
	Inefficiencies  []InefficiencyItem `json:"inefficiencies"`
}

type InefficiencyItem struct {
	Path           string `json:"path"`
	CumulativeSize int64  `json:"cumulative_size_bytes"`
	Occurrences    int    `json:"occurrences"`
}

type FileNodeInfo struct {
	Path     string `json:"path"`
	Type     string `json:"type"` // "file" or "directory"
	Size     uint64 `json:"size_bytes"`
	DiffType string `json:"diff_type,omitempty"` // "added", "modified", "removed", "unmodified"
}

type DiffResult struct {
	Image       string         `json:"image"`
	BaseLayer   int            `json:"base_layer_index"`
	TargetLayer int            `json:"target_layer_index"`
	Changes     []FileNodeInfo `json:"changes"`
}

// --- Helper Functions ---

func (h *toolHandlers) validateSandbox(path string) (string, error) {
	if h.opts.Sandbox == "" {
		return path, nil
	}

	absSandbox, err := filepath.Abs(h.opts.Sandbox)
	if err != nil {
		return "", fmt.Errorf("invalid sandbox path: %v", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid image path: %v", err)
	}

	rel, err := filepath.Rel(absSandbox, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("security: path '%s' is outside of sandbox '%s'", path, h.opts.Sandbox)
	}

	return absPath, nil
}

func (h *toolHandlers) getAnalysis(ctx context.Context, imageName string, sourceStr string) (*image.Analysis, error) {
	// Heuristic: if imageName ends in .tar and source is docker, assume docker-archive
	if strings.HasSuffix(imageName, ".tar") && sourceStr == "docker" {
		sourceStr = "docker-archive"
		if _, err := os.Stat(imageName); os.IsNotExist(err) {
			wd, _ := os.Getwd()
			root := wd
			for i := 0; i < 5; i++ {
				if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
					break
				}
				root = filepath.Dir(root)
			}
			dataPath := filepath.Join(root, ".data", imageName)
			if _, err := os.Stat(dataPath); err == nil {
				imageName = dataPath
			}
		}
	}

	source := dive.ParseImageSource(sourceStr)
	if source == dive.SourceUnknown {
		return nil, fmt.Errorf("unknown image source: %s", sourceStr)
	}

	// Security Sandbox check for archives
	if source == dive.SourceDockerArchive {
		var err error
		imageName, err = h.validateSandbox(imageName)
		if err != nil {
			return nil, err
		}
	}

	cacheKey := fmt.Sprintf("%s:%s", sourceStr, imageName)

	if cached, ok := h.analyses.Get(cacheKey); ok {
		ttl, err := time.ParseDuration(h.opts.CacheTTL)
		if err == nil && time.Since(cached.Timestamp) < ttl {
			return cached.Analysis, nil
		}
		h.analyses.Remove(cacheKey)
	}

	log.Infof("Image %s not in cache or expired, analyzing...", imageName)
	resolver, err := dive.GetImageResolver(source)
	if err != nil {
		return nil, fmt.Errorf("cannot get image resolver: %v", err)
	}

	img, err := adapter.ImageResolver(resolver).Fetch(ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch image: %v", err)
	}

	analysis, err := adapter.NewAnalyzer().Analyze(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("cannot analyze image: %v", err)
	}

	h.analyses.Add(cacheKey, cachedAnalysis{
		Analysis:  analysis,
		Timestamp: time.Now(),
	})
	return analysis, nil
}

func (h *toolHandlers) jsonResponse(v interface{}) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(b)), nil
}

// --- Tool Handlers ---

func (h *toolHandlers) analyzeImageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imageName, err := request.RequireString("image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sourceStr := request.GetString("source", "docker")

	analysis, err := h.getAnalysis(ctx, imageName, sourceStr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	summary := ImageSummary{
		Image:           analysis.Image,
		TotalSize:       analysis.SizeBytes,
		EfficiencyScore: analysis.Efficiency,
		WastedSpace:     analysis.WastedBytes,
		LayerCount:      len(analysis.Layers),
		Layers:          make([]LayerSummary, len(analysis.Layers)),
	}

	for i, layer := range analysis.Layers {
		summary.Layers[i] = LayerSummary{
			Index:   i,
			ID:      layer.ShortId(),
			Size:    layer.Size,
			Command: layer.Command,
		}
	}

	return h.jsonResponse(summary)
}

func (h *toolHandlers) getWastedSpaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imageName, err := request.RequireString("image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sourceStr := request.GetString("source", "docker")

	analysis, err := h.getAnalysis(ctx, imageName, sourceStr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := WastedSpaceResult{
		Image:          analysis.Image,
		Inefficiencies: make([]InefficiencyItem, 0),
	}

	limit := 20
	if len(analysis.Inefficiencies) < limit {
		limit = len(analysis.Inefficiencies)
	}

	for i := 0; i < limit; i++ {
		inef := analysis.Inefficiencies[i]
		result.Inefficiencies = append(result.Inefficiencies, InefficiencyItem{
			Path:           inef.Path,
			CumulativeSize: inef.CumulativeSize,
			Occurrences:    len(inef.Nodes),
		})
	}

	return h.jsonResponse(result)
}

func (h *toolHandlers) inspectLayerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imageName, err := request.RequireString("image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	layerIdx, err := request.RequireInt("layer_index")
	if err != nil {
		return mcp.NewToolResultError("layer_index is required and must be an integer"), nil
	}

	sourceStr := request.GetString("source", "docker")
	pathStr := request.GetString("path", "/")

	analysis, err := h.getAnalysis(ctx, imageName, sourceStr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if layerIdx < 0 || layerIdx >= len(analysis.RefTrees) {
		return mcp.NewToolResultError(fmt.Sprintf("layer index out of bounds: %d (total layers: %d)", layerIdx, len(analysis.RefTrees))), nil
	}

	tree := analysis.RefTrees[layerIdx]
	
	startNode := tree.Root
	if pathStr != "/" {
		node, err := tree.GetNode(pathStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("path not found in layer: %s", pathStr)), nil
		}
		startNode = node
	}

	files := make([]FileNodeInfo, 0)
	count := 0
	limit := 500 // Higher limit for JSON

	for name, child := range startNode.Children {
		if count >= limit {
			break
		}
		typeStr := "file"
		if child.Data.FileInfo.IsDir {
			typeStr = "directory"
		}
		files = append(files, FileNodeInfo{
			Path: filepath.Join(pathStr, name),
			Type: typeStr,
			Size: uint64(child.Data.FileInfo.Size),
		})
		count++
	}

	return h.jsonResponse(files)
}

func (h *toolHandlers) diffLayersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imageName, err := request.RequireString("image")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	baseIdx, err := request.RequireInt("base_layer_index")
	if err != nil {
		return mcp.NewToolResultError("base_layer_index is required"), nil
	}

	targetIdx, err := request.RequireInt("target_layer_index")
	if err != nil {
		return mcp.NewToolResultError("target_layer_index is required"), nil
	}

	sourceStr := request.GetString("source", "docker")

	analysis, err := h.getAnalysis(ctx, imageName, sourceStr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if baseIdx < 0 || baseIdx >= len(analysis.RefTrees) || targetIdx < 0 || targetIdx >= len(analysis.RefTrees) {
		return mcp.NewToolResultError("layer index out of bounds"), nil
	}

	baseTree, _, err := filetree.StackTreeRange(analysis.RefTrees, 0, baseIdx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stack base tree: %v", err)), nil
	}

	targetTree, _, err := filetree.StackTreeRange(analysis.RefTrees, 0, targetIdx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to stack target tree: %v", err)), nil
	}

	_, err = baseTree.CompareAndMark(targetTree)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to compare trees: %v", err)), nil
	}

	changes := make([]FileNodeInfo, 0)
	
	err = baseTree.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.DiffType != filetree.Unmodified {
			diffStr := ""
			switch node.Data.DiffType {
			case filetree.Added:
				diffStr = "added"
			case filetree.Modified:
				diffStr = "modified"
			case filetree.Removed:
				diffStr = "removed"
			}
			
			typeStr := "file"
			if node.Data.FileInfo.IsDir {
				typeStr = "directory"
			}

			changes = append(changes, FileNodeInfo{
				Path:     node.Path(),
				Type:     typeStr,
				Size:     uint64(node.Data.FileInfo.Size),
				DiffType: diffStr,
			})
		}
		return nil
	}, nil)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to visit tree: %v", err)), nil
	}

	result := DiffResult{
		Image:       imageName,
		BaseLayer:   baseIdx,
		TargetLayer: targetIdx,
		Changes:     changes,
	}

	return h.jsonResponse(result)
}

// --- Resource Handlers ---

func (h *toolHandlers) resourceSummaryHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	parts := strings.Split(request.Params.URI, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid resource URI: %s", request.Params.URI)
	}
	imageName := parts[3]
	
	analysis, err := h.getAnalysis(ctx, imageName, "docker")
	if err != nil {
		return nil, err
	}

	summary := ImageSummary{
		Image:           analysis.Image,
		TotalSize:       analysis.SizeBytes,
		EfficiencyScore: analysis.Efficiency,
		WastedSpace:     analysis.WastedBytes,
		LayerCount:      len(analysis.Layers),
		Layers:          make([]LayerSummary, len(analysis.Layers)),
	}

	for i, layer := range analysis.Layers {
		summary.Layers[i] = LayerSummary{
			Index:   i,
			ID:      layer.ShortId(),
			Size:    layer.Size,
			Command: layer.Command,
		}
	}

	b, _ := json.MarshalIndent(summary, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(b),
		},
	}, nil
}

func (h *toolHandlers) resourceEfficiencyHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	parts := strings.Split(request.Params.URI, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid resource URI: %s", request.Params.URI)
	}
	imageName := parts[3]
	
	analysis, err := h.getAnalysis(ctx, imageName, "docker")
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"image":            analysis.Image,
		"efficiency_score": analysis.Efficiency,
		"wasted_bytes":     analysis.WastedBytes,
	}

	b, _ := json.MarshalIndent(result, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(b),
		},
	}, nil
}

// --- Prompt Handlers ---

func (h *toolHandlers) promptOptimizeDockerfileHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	imageName, ok := request.Params.Arguments["image"]
	if !ok {
		return nil, fmt.Errorf("image argument is required")
	}

	analysis, err := h.getAnalysis(ctx, imageName, "docker")
	if err != nil {
		return nil, err
	}

	wastedB, _ := json.MarshalIndent(analysis.Inefficiencies, "", "  ")
	summaryB, _ := json.MarshalIndent(analysis, "", "  ")

	return &mcp.GetPromptResult{
		Description: "Optimize Dockerfile based on Dive analysis",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("You are an expert in Docker and OCI image optimization. Your findings for image '%s':\n\nSummary:\n%s\n\nWasted Space:\n%s\n\nPlease suggest optimizations for the Dockerfile.", imageName, string(summaryB), string(wastedB)),
				},
			},
		},
	}, nil
}

func (h *toolHandlers) promptExplainLayerHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	imageName, ok := request.Params.Arguments["image"]
	if !ok {
		return nil, fmt.Errorf("image argument is required")
	}
	layerIdxStr, ok := request.Params.Arguments["layer_index"]
	if !ok {
		return nil, fmt.Errorf("layer_index argument is required")
	}

	return &mcp.GetPromptResult{
		Description: "Explain the impact of a specific image layer",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Can you explain what is happening in layer %s of image '%s'?", layerIdxStr, imageName),
				},
			},
		},
	}, nil
}
