package filetree

import (
	"maps"
	"slices"
	"sort"
)

type SortOrder int

const (
	ByName = iota
	BySizeDesc

	NumSortOrderConventions
)

type OrderStrategy interface {
	orderKeys(files map[string]*FileNode) []string
}

func GetSortOrderStrategy(sortOrder SortOrder) OrderStrategy {
	switch sortOrder {
	case ByName:
		return orderByNameStrategy{}
	case BySizeDesc:
		return orderBySizeDescStrategy{}
	}
	return orderByNameStrategy{}
}

type orderByNameStrategy struct{}

func (orderByNameStrategy) orderKeys(files map[string]*FileNode) []string {
	keys := slices.Collect(maps.Keys(files))

	sort.Strings(keys)

	return keys
}

type orderBySizeDescStrategy struct{}

func (orderBySizeDescStrategy) orderKeys(files map[string]*FileNode) []string {
	keys := slices.Collect(maps.Keys(files))

	sort.Slice(keys, func(i, j int) bool {
		ki, kj := keys[i], keys[j]
		ni, nj := files[ki], files[kj]
		if ni.GetSize() == nj.GetSize() {
			return ki < kj
		}
		return ni.GetSize() > nj.GetSize()
	})

	return keys
}
