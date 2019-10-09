package filetree

import (
	"github.com/sirupsen/logrus"
)

type TreeCacheKey struct {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int
}

type TreeCache struct {
	refTrees []*FileTree
	cache    map[TreeCacheKey]*FileTree
}

func (cache *TreeCache) Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) (*FileTree, error) {
	key := TreeCacheKey{bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop}
	if value, exists := cache.cache[key]; exists {
		return value, nil
	}

	value, err := cache.buildTree(key)
	if err != nil {
		return nil, err
	}
	cache.cache[key] = value
	return value, nil
}

func (cache *TreeCache) buildTree(key TreeCacheKey) (*FileTree, error) {
	newTree, err := StackTreeRange(cache.refTrees, key.bottomTreeStart, key.bottomTreeStop)
	if err != nil {
		return nil, err
	}
	for idx := key.topTreeStart; idx <= key.topTreeStop; idx++ {
		err := newTree.CompareAndMark(cache.refTrees[idx])
		if err != nil {
			logrus.Errorf("unable to build tree: %+v", err)
			return nil, err
		}
	}
	return newTree, nil
}

func (cache *TreeCache) Build() error {
	var bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int

	// case 1: layer compare (top tree SIZE is fixed (BUT floats forward), Bottom tree SIZE changes)
	for selectIdx := 0; selectIdx < len(cache.refTrees); selectIdx++ {
		bottomTreeStart = 0
		topTreeStop = selectIdx

		if selectIdx == 0 {
			bottomTreeStop = selectIdx
			topTreeStart = selectIdx
		} else {
			bottomTreeStop = selectIdx - 1
			topTreeStart = selectIdx
		}

		_, err := cache.Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
		if err != nil {
			return err
		}
	}

	// case 2: aggregated compare (bottom tree is ENTIRELY fixed, top tree SIZE changes)
	for selectIdx := 0; selectIdx < len(cache.refTrees); selectIdx++ {
		bottomTreeStart = 0
		topTreeStop = selectIdx
		if selectIdx == 0 {
			bottomTreeStop = selectIdx
			topTreeStart = selectIdx
		} else {
			bottomTreeStop = 0
			topTreeStart = 1
		}

		_, err := cache.Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewFileTreeCache(refTrees []*FileTree) TreeCache {

	return TreeCache{
		refTrees: refTrees,
		cache:    make(map[TreeCacheKey]*FileTree),
	}
}
