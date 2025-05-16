package filetree

// ViewInfo contains UI specific detail for a specific FileNode
type ViewInfo struct {
	Collapsed bool
	Hidden    bool
}

// NewViewInfo creates a default ViewInfo
func NewViewInfo() (view *ViewInfo) {
	return &ViewInfo{
		Collapsed: GlobalFileTreeCollapse,
		Hidden:    false,
	}
}

// Copy duplicates a ViewInfo
func (view *ViewInfo) Copy() (newView *ViewInfo) {
	newView = NewViewInfo()
	*newView = *view
	return newView
}
