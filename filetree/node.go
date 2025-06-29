// FileNode represents a single file, its attributes, and its position in the file tree
type FileNode struct {
	Tree             *FileTree
	Parent           *FileNode
	Name             string
	Data             FileInfo
	Path             string
	Children         map[string]*FileNode
	AttrAvgSize      uint64
	ViewInfo         *ViewInfo
	area             image.Rectangle
	reversePath      string
	pathHash         string
	formatting       TextFormatting
	RefNodeStatCache map[string]struct{}
	XAttrs           map[string][]byte // Extended attributes (e.g., POSIX capabilities)
}
	node.XAttrs = make(map[string][]byte)
