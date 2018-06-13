package filetree

import (
	"sort"
	"strings"

	"github.com/fatih/color"
)

type FileNode struct {
	Tree      *FileTree
	Parent    *FileNode
	Name      string
	Collapsed bool
	Hidden    bool
	Data      *FileChangeInfo
	Children  map[string]*FileNode
}

func NewNode(parent *FileNode, name string, data *FileChangeInfo) (node *FileNode) {
	node = new(FileNode)
	node.Name = name
	if data == nil {
		data = &FileChangeInfo{}
	}
	node.Data = data
	node.Children = make(map[string]*FileNode)
	node.Parent = parent
	if parent != nil {
		node.Tree = parent.Tree
	}
	return node
}

func (node *FileNode) Copy() *FileNode {
	// newNode := new(FileNode)
	// *newNode = *node
	// return newNode
	newNode := NewNode(node.Parent, node.Name, node.Data)
	for name, child := range node.Children {
		newNode.Children[name] = child.Copy()
	}
	return newNode
}

func (node *FileNode) AddChild(name string, data *FileChangeInfo) (child *FileNode) {
	child = NewNode(node, name, data)
	if node.Children[name] != nil {
		// tree node already exists, replace the payload, keep the children
		node.Children[name].Data = data
	} else {
		node.Children[name] = child
		node.Tree.Size++
	}
	return child
}

func (node *FileNode) Remove() error {
	for _, child := range node.Children {
		child.Remove()
	}
	delete(node.Parent.Children, node.Name)
	node.Tree.Size--
	return nil
}

func (node *FileNode) String() string {
	var style *color.Color
	if node == nil {
		return ""
	} else if node.Data == nil {
		return node.Name
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
	return style.Sprint(node.Name)
}

func (node *FileNode) Visit(visiter Visiter) error {
	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		err := child.Visit(visiter)
		if err != nil {
			return err
		}
	}
	return visiter(node)
}

func (node *FileNode) VisitDepthParentFirst(visiter Visiter, evaluator VisitEvaluator) error {
	err := visiter(node)
	if err != nil {
		return err
	}

	var keys []string
	for key := range node.Children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, name := range keys {
		child := node.Children[name]
		if evaluator == nil || !evaluator(node) {
			continue
		}
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
		node.AssignDiffType(diffType)
		return nil
	}
	myDiffType := diffType

	for _, v := range node.Children {
		vData := v.Data
		myDiffType = myDiffType.merge(vData.DiffType)

	}
	node.AssignDiffType(myDiffType)
	return nil
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

	return a.Data.getDiffType(b.Data)
}
