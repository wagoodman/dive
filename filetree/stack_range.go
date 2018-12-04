package filetree

import (
	"github.com/sirupsen/logrus"
)

// type TreeStackCacheKey struct {
// 	start, stop int
// }
//
// type TreeStackStack map[TreeStackCacheKey]*FileTree
//
// var treeStackCache = make(TreeStackStack)

// StackTreeRange combines an array of trees into a single tree
func StackTreeRange(trees []*FileTree, start, stop int) *FileTree {
	// key := TreeStackCacheKey{start, stop}
	// if value, exists := treeStackCache[key]; exists {
	// 	return value.Copy()
	// }

	tree := trees[0].Copy()
	for idx := start; idx <= stop; idx++ {
		err := tree.Stack(trees[idx])
		if err != nil {
			logrus.Debug("could not stack tree range:", err)
		}
	}
	// treeStackCache[key] = tree
	return tree
}

