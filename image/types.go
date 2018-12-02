package image

import (
	"github.com/docker/docker/client"
	"github.com/wagoodman/dive/filetree"
)

type Parser interface {
}

type Analyzer interface {
	Parse(id string) error
	Analyze() (*AnalysisResult, error)
}

type Layer interface {
	Id() string
	ShortId() string
	Index() int
	Command() string
	Size() uint64
	Tree() *filetree.FileTree
	String() string
}

type AnalysisResult struct {
	Layers         []Layer
	RefTrees       []*filetree.FileTree
	Efficiency     float64
	Inefficiencies filetree.EfficiencySlice
}

type dockerImageAnalyzer struct {
	id        string
	client    *client.Client
	jsonFiles map[string][]byte
	trees     []*filetree.FileTree
	layerMap  map[string]*filetree.FileTree
	layers    []*dockerLayer
}

type dockerImageHistoryEntry struct {
	ID         string
	Size       uint64
	Created    string `json:"created"`
	Author     string `json:"author"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

type dockerImageManifest struct {
	ConfigPath    string   `json:"Config"`
	RepoTags      []string `json:"RepoTags"`
	LayerTarPaths []string `json:"Layers"`
}

type dockerImageConfig struct {
	History []dockerImageHistoryEntry `json:"history"`
	RootFs  dockerRootFs              `json:"rootfs"`
}

type dockerRootFs struct {
	Type    string   `json:"type"`
	DiffIds []string `json:"diff_ids"`
}

// Layer represents a Docker image layer and metadata
type dockerLayer struct {
	tarPath string
	history dockerImageHistoryEntry
	index   int
	tree    *filetree.FileTree
}
