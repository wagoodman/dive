package filetree

import (
	"fmt"
	"strings"
	"github.com/satori/go.uuid"
)

const (
	newLine         = "\n"
	noBranchSpace   = "    "
	branchSpace     = "│   "
	middleItem      = "├─"
	lastItem        = "└─"
	whiteoutPrefix  = ".wh."
	uncollapsedItem = "─ "
	collapsedItem   = "⊕ "
)

type FileTree struct {
	Root *FileNode
	Size int
	FileSize int64
	Name string
	Id   uuid.UUID
}

func NewFileTree() (tree *FileTree) {
	tree = new(FileTree)
	tree.Size = 0
	tree.Root = new(FileNode)
	tree.Root.Tree = tree
	tree.Root.Children = make(map[string]*FileNode)
	tree.Id = uuid.Must(uuid.NewV4())
	return tree
}

func (tree *FileTree) String(showAttributes bool) string {
	return tree.Root.renderStringTree([]bool{}, showAttributes, 0)
}

func (tree *FileTree) Copy() *FileTree {
	newTree := NewFileTree()
	newTree.Size = tree.Size
	newTree.FileSize = tree.FileSize
	newTree.Root = tree.Root.Copy(newTree.Root)

	// update the tree pointers
	newTree.VisitDepthChildFirst(func(node *FileNode) error {
		node.Tree = newTree
		return nil
	}, nil)

	return newTree
}

type Visiter func(*FileNode) error
type VisitEvaluator func(*FileNode) bool

// DFS bubble up
func (tree *FileTree) VisitDepthChildFirst(visiter Visiter, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthChildFirst(visiter, evaluator)
}

// DFS sink down
func (tree *FileTree) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	return tree.Root.VisitDepthParentFirst(visiter, evaluator)
}

func (tree *FileTree) Stack(upper *FileTree) error {
	graft := func(node *FileNode) error {
		if node.IsWhiteout() {
			err := tree.RemovePath(node.Path())
			if err != nil {
				return fmt.Errorf("cannot remove node %s: %v", node.Path(), err.Error())
			}
		} else {
			newNode, err := tree.AddPath(node.Path(), node.Data.FileInfo)
			if err != nil {
				return fmt.Errorf("cannot add node %s: %v", newNode.Path(), err.Error())
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft, nil)
}

func (tree *FileTree) GetNode(path string) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	for _, name := range nodeNames {
		if name == "" {
			continue
		}
		if node.Children[name] == nil {
			return nil, fmt.Errorf("path does not exist: %s", path)
		}
		node = node.Children[name]
	}
	return node, nil
}

func (tree *FileTree) AddPath(path string, data FileInfo) (*FileNode, error) {
	nodeNames := strings.Split(strings.Trim(path, "/"), "/")
	node := tree.Root
	for idx, name := range nodeNames {
		if name == "" {
			continue
		}
		// find or create node
		if node.Children[name] != nil {
			node = node.Children[name]
		} else {
			// don't attach the payload. The payload is destined for the
			// Path's end node, not any intermediary node.
			node = node.AddChild(name, FileInfo{})
		}

		// attach payload to the last specified node
		if idx == len(nodeNames)-1 {
			node.Data.FileInfo = data
		}

	}
	return node, nil
}

func (tree *FileTree) RemovePath(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.Remove()
}

func (tree *FileTree) Compare(upper *FileTree) error {
	graft := func(upperNode *FileNode) error {
		if upperNode.IsWhiteout() {
			err := tree.MarkRemoved(upperNode.Path())
			if err != nil {
				return fmt.Errorf("cannot remove upperNode %s: %v", upperNode.Path(), err.Error())
			}
		} else {
			lowerNode, _ := tree.GetNode(upperNode.Path())
			if lowerNode == nil {
				newNode, err := tree.AddPath(upperNode.Path(), upperNode.Data.FileInfo)
				if err != nil {
					return fmt.Errorf("cannot add new upperNode %s: %v", upperNode.Path(), err.Error())
				}
				newNode.AssignDiffType(Added)
			} else {
				diffType := lowerNode.compare(upperNode)
				return lowerNode.deriveDiffType(diffType)
			}
		}
		return nil
	}
	return upper.VisitDepthChildFirst(graft, nil)
}

func (tree *FileTree) MarkRemoved(path string) error {
	node, err := tree.GetNode(path)
	if err != nil {
		return err
	}
	return node.AssignDiffType(Removed)
}

// memoize StackRange for performance
type stackRangeCacheKey struct {
	// Ids mapset.Set
	start, stop int
}

var stackRangeCache = make(map[stackRangeCacheKey]*FileTree)

func StackRange(trees []*FileTree, start, stop int) *FileTree {

	// var ids []interface{}
	//
	// for _, tree := range trees {
	// 	ids = append(ids, tree.Id)
	// }
//mapset.NewSetFromSlice(ids)
// 	key := stackRangeCacheKey{start, stop}
//
//
// 	cachedResult, ok := stackRangeCache[key]
// 	if ok {
// 		return cachedResult
// 	}

	tree := trees[0].Copy()
	for idx := start; idx <= stop; idx++ {
		tree.Stack(trees[idx])
	}

	// stackRangeCache[key] = tree

	return tree
}

// EfficiencyMap creates a map[string]int showing how often each int
// appears in the
func EfficiencyMap(trees []*FileTree) map[string]int {
	result := make(map[string]int)
	visitor := func(node *FileNode) error {
		result[node.Path()]++
		return nil
	}
	visitEvaluator := func(node *FileNode) bool {
		return node.IsLeaf()
	}
	for _, tree := range trees {
		tree.VisitDepthChildFirst(visitor, visitEvaluator)
	}
	return result
}

func EfficiencyScore(trees []*FileTree) float64 {
	efficiencyMap := EfficiencyMap(trees)
	uniquePaths := len(efficiencyMap)
	pathAppearances := 0
	for _, value := range efficiencyMap {
		pathAppearances += value
	}
	return float64(uniquePaths) / float64(pathAppearances)
}
