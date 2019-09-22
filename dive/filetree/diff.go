package filetree

import (
	"fmt"
)

const (
	Unmodified DiffType = iota
	Modified
	Added
	Removed
)

// DiffType defines the comparison result between two FileNodes
type DiffType int

// String of a DiffType
func (diff DiffType) String() string {
	switch diff {
	case Unmodified:
		return "Unmodified"
	case Modified:
		return "Modified"
	case Added:
		return "Added"
	case Removed:
		return "Removed"
	default:
		return fmt.Sprintf("%d", int(diff))
	}
}

// merge two DiffTypes into a single result. Essentially, return the given value unless they two values differ,
// in which case we can only determine that there is "a change".
func (diff DiffType) merge(other DiffType) DiffType {
	if diff == other {
		return diff
	}
	return Modified
}
