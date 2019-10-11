package ui

type layoutManager struct {
	fileTreeSplitRatio float64
}

func newLayoutManager(fileTreeSplitRatio float64) *layoutManager {
	return &layoutManager{
		fileTreeSplitRatio: fileTreeSplitRatio,
	}
}
