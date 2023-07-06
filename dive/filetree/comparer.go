package filetree

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type TreeIndexKey struct {
	bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int
}

func NewTreeIndexKey(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) TreeIndexKey {
	return TreeIndexKey{
		bottomTreeStart: bottomTreeStart,
		bottomTreeStop:  bottomTreeStop,
		topTreeStart:    topTreeStart,
		topTreeStop:     topTreeStop,
	}
}

func (index TreeIndexKey) String() string {
	if index.bottomTreeStart == index.bottomTreeStop && index.topTreeStart == index.topTreeStop {
		return fmt.Sprintf("Index(%d:%d)", index.bottomTreeStart, index.topTreeStart)
	} else if index.bottomTreeStart == index.bottomTreeStop {
		return fmt.Sprintf("Index(%d:%d-%d)", index.bottomTreeStart, index.topTreeStart, index.topTreeStop)
	} else if index.topTreeStart == index.topTreeStop {
		return fmt.Sprintf("Index(%d-%d:%d)", index.bottomTreeStart, index.bottomTreeStop, index.topTreeStart)
	}
	return fmt.Sprintf("Index(%d-%d:%d-%d)", index.bottomTreeStart, index.bottomTreeStop, index.topTreeStart, index.topTreeStop)
}

type Comparer struct {
	refTrees   []*FileTree
	trees      map[TreeIndexKey]*FileTree
	pathErrors map[TreeIndexKey][]PathError
}

func NewComparer(refTrees []*FileTree) Comparer {
	return Comparer{
		refTrees:   refTrees,
		trees:      make(map[TreeIndexKey]*FileTree),
		pathErrors: make(map[TreeIndexKey][]PathError),
	}
}

func (cmp *Comparer) GetPathErrors(key TreeIndexKey) ([]PathError, error) {
	_, pathErrors, err := cmp.get(key)
	if err != nil {
		return nil, err
	}
	return pathErrors, nil
}

func (cmp *Comparer) GetTree(key TreeIndexKey) (*FileTree, error) {
	// func (cmp *Comparer) GetTree(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) (*FileTree, []PathError, error) {
	// key := TreeIndexKey{bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop}

	if value, exists := cmp.trees[key]; exists {
		return value, nil
	}

	value, pathErrors, err := cmp.get(key)
	if err != nil {
		return nil, err
	}
	cmp.trees[key] = value
	cmp.pathErrors[key] = pathErrors
	return value, nil
}

func (cmp *Comparer) get(key TreeIndexKey) (*FileTree, []PathError, error) {
	newTree, pathErrors, err := StackTreeRange(cmp.refTrees, key.bottomTreeStart, key.bottomTreeStop)
	if err != nil {
		return nil, nil, err
	}
	for idx := key.topTreeStart; idx <= key.topTreeStop; idx++ {
		markPathErrors, err := newTree.CompareAndMark(cmp.refTrees[idx])
		pathErrors = append(pathErrors, markPathErrors...)
		if err != nil {
			logrus.Errorf("error while building tree: %+v", err)
			return nil, nil, err
		}
	}
	return newTree, pathErrors, nil
}

// case 1: layer compare (top tree SIZE is fixed (BUT floats forward), Bottom tree SIZE changes)
func (cmp *Comparer) NaturalIndexes() <-chan TreeIndexKey {
	indexes := make(chan TreeIndexKey)

	go func() {
		defer close(indexes)

		var bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int

		for selectIdx := 0; selectIdx < len(cmp.refTrees); selectIdx++ {
			bottomTreeStart = 0
			topTreeStop = selectIdx

			if selectIdx == 0 {
				bottomTreeStop = selectIdx
				topTreeStart = selectIdx
			} else {
				bottomTreeStop = selectIdx - 1
				topTreeStart = selectIdx
			}

			indexes <- TreeIndexKey{
				bottomTreeStart: bottomTreeStart,
				bottomTreeStop:  bottomTreeStop,
				topTreeStart:    topTreeStart,
				topTreeStop:     topTreeStop,
			}
		}
	}()
	return indexes
}

// case 2: aggregated compare (bottom tree is ENTIRELY fixed, top tree SIZE changes)
func (cmp *Comparer) AggregatedIndexes() <-chan TreeIndexKey {
	indexes := make(chan TreeIndexKey)

	go func() {
		defer close(indexes)

		var bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int

		for selectIdx := 0; selectIdx < len(cmp.refTrees); selectIdx++ {
			bottomTreeStart = 0
			topTreeStop = selectIdx
			if selectIdx == 0 {
				bottomTreeStop = selectIdx
				topTreeStart = selectIdx
			} else {
				bottomTreeStop = 0
				topTreeStart = 1
			}

			indexes <- TreeIndexKey{
				bottomTreeStart: bottomTreeStart,
				bottomTreeStop:  bottomTreeStop,
				topTreeStart:    topTreeStart,
				topTreeStop:     topTreeStop,
			}
		}
	}()
	return indexes
}

func (cmp *Comparer) BuildCache() (errors []error) {
	for index := range cmp.NaturalIndexes() {
		pathError, _ := cmp.GetPathErrors(index)
		if len(pathError) > 0 {
			for _, path := range pathError {
				errors = append(errors, fmt.Errorf("path error at layer index %s: %s", index, path))
			}
		}
		_, err := cmp.GetTree(index)
		if err != nil {
			errors = append(errors, err)
			return errors
		}
	}

	for index := range cmp.AggregatedIndexes() {
		_, err := cmp.GetTree(index)
		if err != nil {
			errors = append(errors, err)
			return errors
		}
	}
	return errors
}
