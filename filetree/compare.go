// Compare indicates if a node has changes relative to another node.
func (node *FileNode) Compare(other *FileNode) DiffType {
	// don't compare dir stats, we dont capture that by default
	if other == nil || node.Data.IsDir != other.Data.IsDir {
		return Modified
	}

	if node.Data.IsDir {
		return Unmodified
	}

	if node.Data.Size != other.Data.Size {
		return Modified
	}

	if node.Data.Mode != other.Data.Mode {
		return Modified
	}

	if node.Data.MD5sum != other.Data.MD5sum {
		return Modified
	}

	// Check extended attributes for differences
	if !compareXAttrs(node.Data.XAttrs, other.Data.XAttrs) {
		return Modified
	}

	return Unmodified
}

// compareXAttrs checks if two XAttrs maps are equal
func compareXAttrs(a, b map[string][]byte) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		if !bytes.Equal(v1, v2) {
			return false
		}
	}

	return true
}
