package filetree

import (
	"sort"
)

type EfficiencySlice []*EfficiencyData

type EfficiencyData struct {
	Path              string
	Nodes             []*FileNode
	CumulativeSize    int64
	minDiscoveredSize int64
}

func (d EfficiencySlice) Len() int {
	return len(d)
}

func (d EfficiencySlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d EfficiencySlice) Less(i, j int) bool {
	return d[i].CumulativeSize < d[j].CumulativeSize
}


// 1. Files that are duplicated across layers discounts your score, weighted by file size
// 2. Files that are removed discounts your score, weighted by the original file size
func Efficiency(trees []*FileTree) (float64, EfficiencySlice) {
	efficiencyMap := make(map[string]*EfficiencyData)
	inefficientMatches := make(EfficiencySlice, 0)

	visitor := func(node *FileNode) error {
		path := node.Path()
		if _, ok := efficiencyMap[path]; !ok {
			efficiencyMap[path] = &EfficiencyData{
				Path:              path,
				Nodes:             make([]*FileNode,0),
				minDiscoveredSize: -1,
			}
		}
		data := efficiencyMap[path]
		data.CumulativeSize += node.Data.FileInfo.TarHeader.Size
		if data.minDiscoveredSize < 0 || node.Data.FileInfo.TarHeader.Size < data.minDiscoveredSize {
			data.minDiscoveredSize = node.Data.FileInfo.TarHeader.Size
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
	for _, tree := range trees {
		tree.VisitDepthChildFirst(visitor, visitEvaluator)
	}


	// calculate the score
	var minimumPathSizes int64
	var discoveredPathSizes int64

	for _, value := range efficiencyMap {
		minimumPathSizes += value.minDiscoveredSize
		discoveredPathSizes += value.CumulativeSize
	}
	score := float64(minimumPathSizes) / float64(discoveredPathSizes)

	sort.Sort(inefficientMatches)

	return score, inefficientMatches
}


