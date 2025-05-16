package types

import (
	"github.com/wagoodman/dive/dive/v1/filetree"
)

// Visitor is a function that processes, observes, or otherwise transforms the given node
type Visitor func(FileNode) error

// VisitEvaluator is a function that indicates whether the given node should be visited by a Visitor.
type VisitEvaluator func(FileNode) bool

type FileNode interface {
	Path() string
	Parent() FileNode
	ViewInfo() filetree.ViewInfo
	FileInfo() filetree.FileInfo
	DiffType() filetree.DiffType
	Children() map[string]*FileNode
}

type FileTree interface {
	VisibleSize() int
	String(showAttributes bool) string
	StringBetween(start, stop int, showAttributes bool) string
	Copy() FileTree
	VisitDepthParentFirst(visitor Visitor, evaluator VisitEvaluator) error
	VisitDepthChildFirst(visitor Visitor, evaluator VisitEvaluator) error
	SortOrder() filetree.SortOrder
	RemovePath(path string) error
}
