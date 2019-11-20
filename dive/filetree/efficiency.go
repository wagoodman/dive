package filetree

import (
	"sort"

	"github.com/sirupsen/logrus"
)

// EfficiencyData represents the storage and reference statistics for a given file tree path.
type EfficiencyData struct {
	Path              string
	Nodes             []*FileNode
	CumulativeSize    int64
	minDiscoveredSize int64
}

// EfficiencySlice represents an ordered set of EfficiencyData data structures.
type EfficiencySlice []*EfficiencyData

// Len is required for sorting.
func (efs EfficiencySlice) Len() int {
	return len(efs)
}

// Swap operation is required for sorting.
func (efs EfficiencySlice) Swap(i, j int) {
	efs[i], efs[j] = efs[j], efs[i]
}

// Less comparison is required for sorting.
func (efs EfficiencySlice) Less(i, j int) bool {
	return efs[i].CumulativeSize < efs[j].CumulativeSize
}

// Efficiency returns the score and file set of the given set of FileTrees (layers). This is loosely based on:
// 1. Files that are duplicated across layers discounts your score, weighted by file size
// 2. Files that are removed discounts your score, weighted by the original file size
func Efficiency(trees []*FileTree) (float64, EfficiencySlice) {
	efficiencyMap := make(map[string]*EfficiencyData)
	inefficientMatches := make(EfficiencySlice, 0)
	currentTree := 0

	visitor := func(node *FileNode) error {
		path := node.Path()
		if _, ok := efficiencyMap[path]; !ok {
			efficiencyMap[path] = &EfficiencyData{
				Path:              path,
				Nodes:             make([]*FileNode, 0),
				minDiscoveredSize: -1,
			}
		}
		data := efficiencyMap[path]

		// this node may have had children that were deleted, however, we won't explicitly list out every child, only
		// the top-most parent with the cumulative size. These operations will need to be done on the full (stacked)
		// tree.
		// Note: whiteout files may also represent directories, so we need to find out if this was previously a file or dir.
		var sizeBytes int64

		if node.IsWhiteout() {
			sizer := func(curNode *FileNode) error {
				sizeBytes += curNode.Data.FileInfo.Size
				return nil
			}
			stackedTree, failedPaths, err := StackTreeRange(trees, 0, currentTree-1)
			if len(failedPaths) > 0 {
				for _, path := range failedPaths {
					logrus.Errorf(path.String())
				}
			}
			if err != nil {
				logrus.Errorf("unable to stack tree range: %+v", err)
				return err
			}

			previousTreeNode, err := stackedTree.GetNode(node.Path())
			if err != nil {
				return err
			}

			if previousTreeNode.Data.FileInfo.IsDir {
				err = previousTreeNode.VisitDepthChildFirst(sizer, nil)
				if err != nil {
					logrus.Errorf("unable to propagate whiteout dir: %+v", err)
					return err
				}
			}

		} else {
			sizeBytes = node.Data.FileInfo.Size
		}

		data.CumulativeSize += sizeBytes
		if data.minDiscoveredSize < 0 || sizeBytes < data.minDiscoveredSize {
			data.minDiscoveredSize = sizeBytes
		}
		data.Nodes = append(data.Nodes, node)

		if len(data.Nodes) == 2 {
			inefficientMatches = append(inefficientMatches, data)
		}

		return nil
	}
	visitEvaluator := func(node *FileNode) bool {
		return node.IsLeaf()
	}
	for idx, tree := range trees {
		currentTree = idx
		err := tree.VisitDepthChildFirst(visitor, visitEvaluator)
		if err != nil {
			logrus.Errorf("unable to propagate ref tree: %+v", err)
		}
	}

	// calculate the score
	var minimumPathSizes int64
	var discoveredPathSizes int64

	for _, value := range efficiencyMap {
		minimumPathSizes += value.minDiscoveredSize
		discoveredPathSizes += value.CumulativeSize
	}
	var score float64
	if discoveredPathSizes == 0 {
		score = 1.0
	} else {
		score = float64(minimumPathSizes) / float64(discoveredPathSizes)
	}

	sort.Sort(inefficientMatches)

	return score, inefficientMatches
}
