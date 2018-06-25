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

type FileNode struct {
	Tree     *FileTree
	Parent   *FileNode
	Name     string
	Data     NodeData
	Children map[string]*FileNode
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
	var style *color.Color
	var display string
	if node == nil {
		return ""
	}
	switch node.Data.DiffType {
	case Added:
		style = color.New(color.FgGreen)
	case Removed:
		style = color.New(color.FgRed)
	case Changed:
		style = color.New(color.FgYellow)
	case Unchanged:
		style = color.New(color.Reset)
	default:
		style = color.New(color.BgMagenta)
	}
	display = node.Name
	if node.Data.FileInfo.TarHeader.Typeflag == tar.TypeSymlink || node.Data.FileInfo.TarHeader.Typeflag == tar.TypeLink {
		display += " -> " + node.Data.FileInfo.TarHeader.Linkname
	}
	return style.Sprint(display)
}

func (node *FileNode) MetadataString() string {
	var style *color.Color
	if node == nil {
		return ""
	}
	switch node.Data.DiffType {
	case Added:
		style = color.New(color.FgGreen)
	case Removed:
		style = color.New(color.FgRed)
	case Changed:
		style = color.New(color.FgYellow)
	case Unchanged:
		style = color.New(color.Reset)
	default:
		style = color.New(color.BgMagenta)
	}

	fileMode := permbits.FileMode(node.Data.FileInfo.TarHeader.FileInfo().Mode()).String()
	dir := "-"
	if node.Data.FileInfo.TarHeader.FileInfo().IsDir() {
		dir = "d"
	}
	user := node.Data.FileInfo.TarHeader.Uid
	group := node.Data.FileInfo.TarHeader.Gid
	userGroup := fmt.Sprintf("%d:%d", user, group)
	size := humanize.Bytes(uint64(node.Data.FileInfo.TarHeader.FileInfo().Size()))

	return style.Sprint(fmt.Sprintf(AttributeFormat, dir, fileMode, userGroup, size))
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
	return "/" + strings.Join(path, "/")
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
