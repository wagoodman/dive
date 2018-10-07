package filetree

import (
	"archive/tar"
	"fmt"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/phayes/permbits"
)

const (
	AttributeFormat = "%s%s %10s %10s "
)

var diffTypeColor = map[DiffType]*color.Color {
	Added: color.New(color.FgGreen),
	Removed: color.New(color.FgRed),
	Changed: color.New(color.FgYellow),
	Unchanged: color.New(color.Reset),
}

type FileNode struct {
	Tree     *FileTree
	Parent   *FileNode
	Name     string
	Data     NodeData
	Children map[string]*FileNode
	path     string
}

func NewNode(parent *FileNode, name string, data FileInfo) (node *FileNode) {
	node = new(FileNode)
	node.Name = name
	node.Data = *NewNodeData()
	node.Data.FileInfo = *data.Copy()

	node.Children = make(map[string]*FileNode)
	node.Parent = parent
	if parent != nil {
		node.Tree = parent.Tree
	}
	return node
}

func (node *FileNode) renderTreeLine(spaces []bool, last bool, collapsed bool) string {
	var otherBranches string
	for _, space := range spaces {
		if space {
			otherBranches += noBranchSpace
		} else {
			otherBranches += branchSpace
		}
	}

	thisBranch := middleItem
	if last {
		thisBranch = lastItem
	}

	collapsedIndicator := uncollapsedItem
	if collapsed {
		collapsedIndicator = collapsedItem
	}

	return otherBranches + thisBranch + collapsedIndicator + node.String() + newLine
}

// todo: until visitor context is implemented, this can't easily be expressed with the existing visitor implementation
func (node *FileNode) renderStringTreeBetween(startRow, stopRow int, currentRow, renderedLines *uint, spaces []bool, showAttributes bool, depth int) string {
	var result string
	var keys []string

	// if we're beyond the range, don't visit this node or subsequent nodes
	if startRow >= 0 && stopRow >= 0 {
		if *currentRow > uint(stopRow) {
			return result
		}
	}

	// always render the nodes consistently (sorted)
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// grab string representation of child nodes
	for idx, name := range keys {

		child := node.Children[name]
		if child.Data.ViewInfo.Hidden {
			continue
		}

		// only keep the results for nodes within the given range
		doRender := true
		if startRow >= 0 && stopRow >= 0 {
			*currentRow++
			if *currentRow < uint(startRow) && *currentRow > uint(stopRow) {
				doRender = false
			} else {
				*renderedLines++
			}
		}

		if doRender {
			last := idx == (len(node.Children) - 1)
			showCollapsed := child.Data.ViewInfo.Collapsed && len(child.Children) > 0
			if showAttributes {
				result += child.MetadataString() + " "
			}
			result += child.renderTreeLine(spaces, last, showCollapsed)

			if len(child.Children) > 0 && !child.Data.ViewInfo.Collapsed {
				spacesChild := append(spaces, last)
				result += child.renderStringTreeBetween(startRow, stopRow, currentRow, renderedLines, spacesChild, showAttributes, depth+1)
			}
		}

	}
	return result
}

func (node *FileNode) renderStringTree(spaces []bool, showAttributes bool, depth int) string {
	return node.renderStringTreeBetween(-1, -1, nil, nil, spaces, showAttributes, depth)
}

func (node *FileNode) Copy(parent *FileNode) *FileNode {
	newNode := NewNode(parent, node.Name, node.Data.FileInfo)
	newNode.Data.ViewInfo = node.Data.ViewInfo
	newNode.Data.DiffType = node.Data.DiffType
	for name, child := range node.Children {
		newNode.Children[name] = child.Copy(newNode)
		child.Parent = newNode
	}
	return newNode
}

func (node *FileNode) AddChild(name string, data FileInfo) (child *FileNode) {
	child = NewNode(node, name, data)
	if node.Children[name] != nil {
		// tree node already exists, replace the payload, keep the children
		node.Children[name].Data.FileInfo = *data.Copy()
	} else {
		node.Children[name] = child
		node.Tree.Size++
	}
	return child
}

func (node *FileNode) Remove() error {
	if node == node.Tree.Root {
		return fmt.Errorf("cannot remove the tree root")
	}
	for _, child := range node.Children {
		child.Remove()
	}
	delete(node.Parent.Children, node.Name)
	node.Tree.Size--
	return nil
}

func (node *FileNode) String() string {
	var display string
	if node == nil {
		return ""
	}

	display = node.Name
	if node.Data.FileInfo.TarHeader.Typeflag == tar.TypeSymlink || node.Data.FileInfo.TarHeader.Typeflag == tar.TypeLink {
		display += " → " + node.Data.FileInfo.TarHeader.Linkname
	}
	return diffTypeColor[node.Data.DiffType].Sprint(display)
}

func (node *FileNode) MetadataString() string {
	if node == nil {
		return ""
	}

	fileMode := permbits.FileMode(node.Data.FileInfo.TarHeader.FileInfo().Mode()).String()
	dir := "-"
	if node.Data.FileInfo.TarHeader.FileInfo().IsDir() {
		dir = "d"
	}
	user := node.Data.FileInfo.TarHeader.Uid
	group := node.Data.FileInfo.TarHeader.Gid
	userGroup := fmt.Sprintf("%d:%d", user, group)

	//size := humanize.Bytes(uint64(node.Data.FileInfo.TarHeader.FileInfo().Size()))
	var sizeBytes int64

	if node.Data.FileInfo.TarHeader.FileInfo().IsDir() {

		sizer := func(curNode *FileNode) error {
			if curNode.Data.DiffType != Removed {
				sizeBytes += curNode.Data.FileInfo.TarHeader.FileInfo().Size()
			}
			return nil
		}

		node.VisitDepthChildFirst(sizer, nil)
	} else {
		sizeBytes = node.Data.FileInfo.TarHeader.FileInfo().Size()
	}

	size := humanize.Bytes(uint64(sizeBytes))

	return diffTypeColor[node.Data.DiffType].Sprint(fmt.Sprintf(AttributeFormat, dir, fileMode, userGroup, size))
}

func (node *FileNode) VisitDepthChildFirst(visiter Visiter, evaluator VisitEvaluator) error {
	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		err := child.VisitDepthChildFirst(visiter, evaluator)
		if err != nil {
			return err
		}
	}
	// never visit the root node
	if node == node.Tree.Root {
		return nil
	} else if evaluator != nil && evaluator(node) || evaluator == nil {
		return visiter(node)
	}

	return nil
}

func (node *FileNode) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	var err error

	doVisit := evaluator != nil && evaluator(node) || evaluator == nil

	if !doVisit {
		return nil
	}

	// never visit the root node
	if node != node.Tree.Root {
		err = visiter(node)
		if err != nil {
			return err
		}
	}

	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		err = child.VisitDepthParentFirst(visiter, evaluator)
		if err != nil {
			return err
		}
	}
	return err
}

func (node *FileNode) IsWhiteout() bool {
	return strings.HasPrefix(node.Name, whiteoutPrefix)
}

func (node *FileNode) Path() string {
	if node.path == "" {
		path := []string{}
		curNode := node
		for {
			if curNode.Parent == nil {
				break
			}

			name := curNode.Name
			if curNode == node {
				// white out prefixes are fictitious on leaf nodes
				name = strings.TrimPrefix(name, whiteoutPrefix)
			}

			path = append([]string{name}, path...)
			curNode = curNode.Parent
		}
		node.path = "/" + strings.Join(path, "/")
	}
	return node.path
}

func (node *FileNode) IsLeaf() bool {
	return len(node.Children) == 0
}

func (node *FileNode) deriveDiffType(diffType DiffType) error {
	// THE DIFF_TYPE OF A NODE IS ALWAYS THE DIFF_TYPE OF ITS ATTRIBUTES AND ITS CONTENTS
	// THE CONTENTS ARE THE BYTES OF A FILE OR THE CHILDREN OF A DIRECTORY

	if node.IsLeaf() {
		return node.AssignDiffType(diffType)
	}
	myDiffType := diffType

	for _, v := range node.Children {
		myDiffType = myDiffType.merge(v.Data.DiffType)

	}

	return node.AssignDiffType(myDiffType)
}

func (node *FileNode) AssignDiffType(diffType DiffType) error {
	var err error

	// todo, this is an indicator that the root node approach isn't working
	if node.Path() == "/" {
		return nil
	}

	node.Data.DiffType = diffType

	// if we've removed this node, then all children have been removed as well
	if diffType == Removed {
		for _, child := range node.Children {
			err = child.AssignDiffType(diffType)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *FileNode) compare(b *FileNode) DiffType {
	if a == nil && b == nil {
		return Unchanged
	}
	// a is nil but not b
	if a == nil && b != nil {
		return Added
	}

	// b is nil but not a
	if a != nil && b == nil {
		return Removed
	}

	if b.IsWhiteout() {
		return Removed
	}
	if a.Name != b.Name {
		panic("comparing mismatched nodes")
	}
	// TODO: fails on nil

	return a.Data.FileInfo.getDiffType(b.Data.FileInfo)
}
