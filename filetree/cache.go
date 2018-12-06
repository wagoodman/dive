package filetree

type TreeCacheKey struct {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int
}

type TreeCache struct {
	refTrees []*FileTree
	cache    map[TreeCacheKey]*FileTree
}

func (cache *TreeCache) Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) *FileTree {
	key := TreeCacheKey{bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop}
	if value, exists := cache.cache[key]; exists {
		return value
	} else {

	}
	value := cache.buildTree(key)
	cache.cache[key] = value
	return value
}

func (cache *TreeCache) buildTree(key TreeCacheKey) *FileTree {
	newTree := StackTreeRange(cache.refTrees, key.bottomTreeStart, key.bottomTreeStop)

	for idx := key.topTreeStart; idx <= key.topTreeStop; idx++ {
		newTree.Compare(cache.refTrees[idx])
	}
	return newTree
}

func (cache *TreeCache) Build() {
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

		cache.Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
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

		cache.Get(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	}
}

func NewFileTreeCache(refTrees []*FileTree) TreeCache {

	return TreeCache{
		refTrees: refTrees,
		cache:    make(map[TreeCacheKey]*FileTree),
	}
}
