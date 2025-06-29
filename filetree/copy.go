// Copy clones a FileNode into a new one.
func (node *FileNode) Copy(parent *FileNode, tree *FileTree) *FileNode {
	result := NewFileNode(tree, parent, node.Name, node.Data)
	
	// Copy extended attributes
	for k, v := range node.Data.XAttrs {
		result.Data.XAttrs[k] = v
	}
	
	return result
}
