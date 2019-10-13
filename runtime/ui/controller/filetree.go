package controller

import (
	"fmt"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/lunixbochs/vtclean"
	"github.com/wagoodman/dive/dive/filetree"
)

const (
	CompareLayer CompareType = iota
	CompareAll
)

type CompareType int

// FileTree holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
type FileTree struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	vm     *viewmodel.FileTree

	helpKeys []*key.Binding
}

// NewFileTreeController creates a new view object attached the the global [gocui] screen object.
func NewFileTreeController(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree, cache filetree.TreeCache) (controller *FileTree, err error) {
	controller = new(FileTree)

	// populate main fields
	controller.name = name
	controller.gui = gui
	controller.vm, err = viewmodel.NewFileTreeViewModel(tree, refTrees, cache)
	if err != nil {
		return nil, err
	}

	return controller, err
}

func (c *FileTree) Name() string {
	return c.name
}

func (c *FileTree) AreAttributesVisible() bool {
	return c.vm.ShowAttributes
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (c *FileTree) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	c.view = v
	c.view.Editable = false
	c.view.Wrap = false
	c.view.Frame = false

	c.header = header
	c.header.Editable = false
	c.header.Wrap = false
	c.header.Frame = false

	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.toggle-collapse-dir"},
			OnAction:   c.toggleCollapse,
			Display:    "Collapse dir",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-collapse-all-dir"},
			OnAction:   c.toggleCollapseAll,
			Display:    "Collapse all dir",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-added-files"},
			OnAction:   func() error { return c.toggleShowDiffType(filetree.Added) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[filetree.Added] },
			Display:    "Added",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-removed-files"},
			OnAction:   func() error { return c.toggleShowDiffType(filetree.Removed) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[filetree.Removed] },
			Display:    "Removed",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-modified-files"},
			OnAction:   func() error { return c.toggleShowDiffType(filetree.Modified) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[filetree.Modified] },
			Display:    "Modified",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-unchanged-files", "keybinding.toggle-unmodified-files"},
			OnAction:   func() error { return c.toggleShowDiffType(filetree.Unmodified) },
			IsSelected: func() bool { return !c.vm.HiddenDiffTypes[filetree.Unmodified] },
			Display:    "Unmodified",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-filetree-attributes"},
			OnAction:   c.toggleAttributes,
			IsSelected: func() bool { return c.vm.ShowAttributes },
			Display:    "Attributes",
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   c.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   c.PageDown,
		},
		{
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: c.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: c.CursorUp,
		},
		{
			Key:      gocui.KeyArrowLeft,
			Modifier: gocui.ModNone,
			OnAction: c.CursorLeft,
		},
		{
			Key:      gocui.KeyArrowRight,
			Modifier: gocui.ModNone,
			OnAction: c.CursorRight,
		},
	}

	helpKeys, err := key.GenerateBindings(c.gui, c.name, infos)
	if err != nil {
		return err
	}
	c.helpKeys = helpKeys

	_, height := c.view.Size()
	c.vm.Setup(0, height)
	_ = c.Update()
	_ = c.Render()

	return nil
}

// IsVisible indicates if the file tree view pane is currently initialized
func (c *FileTree) IsVisible() bool {
	return c != nil
}

// ResetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (c *FileTree) resetCursor() {
	_ = c.view.SetCursor(0, 0)
	c.vm.ResetCursor()
}

// SetTreeByLayer populates the view model by stacking the indicated image layer file trees.
func (c *FileTree) setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	err := c.vm.SetTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	if err != nil {
		return err
	}
	// controller.ResetCursor()

	_ = c.Update()
	return c.Render()
}

// CursorDown moves the cursor down and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (c *FileTree) CursorDown() error {
	if c.vm.CursorDown() {
		return c.Render()
	}
	return nil
}

// CursorUp moves the cursor up and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (c *FileTree) CursorUp() error {
	if c.vm.CursorUp() {
		return c.Render()
	}
	return nil
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (c *FileTree) CursorLeft() error {
	err := c.vm.CursorLeft(filterRegex())
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// CursorRight descends into directory expanding it if needed
func (c *FileTree) CursorRight() error {
	err := c.vm.CursorRight(filterRegex())
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// PageDown moves to next page putting the cursor on top
func (c *FileTree) PageDown() error {
	err := c.vm.PageDown()
	if err != nil {
		return err
	}
	return c.Render()
}

// PageUp moves to previous page putting the cursor on top
func (c *FileTree) PageUp() error {
	err := c.vm.PageUp()
	if err != nil {
		return err
	}
	return c.Render()
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
// func (controller *FileTree) getAbsPositionNode() (node *filetree.FileNode) {
// 	return controller.vm.getAbsPositionNode(filterRegex())
// }

// ToggleCollapse will collapse/expand the selected FileNode.
func (c *FileTree) toggleCollapse() error {
	err := c.vm.ToggleCollapse(filterRegex())
	if err != nil {
		return err
	}
	_ = c.Update()
	return c.Render()
}

// ToggleCollapseAll will collapse/expand the all directories.
func (c *FileTree) toggleCollapseAll() error {
	err := c.vm.ToggleCollapseAll()
	if err != nil {
		return err
	}
	if c.vm.CollapseAll {
		c.resetCursor()
	}
	_ = c.Update()
	return c.Render()
}

// ToggleAttributes will show/hide file attributes
func (c *FileTree) toggleAttributes() error {
	err := c.vm.ToggleAttributes()
	if err != nil {
		return err
	}
	// we need to render the changes to the status pane as well (not just this contoller/view)
	return controllers.UpdateAndRender()
}

// ToggleShowDiffType will show/hide the selected DiffType in the filetree pane.
func (c *FileTree) toggleShowDiffType(diffType filetree.DiffType) error {
	c.vm.ToggleShowDiffType(diffType)
	// we need to render the changes to the status pane as well (not just this contoller/view)
	return controllers.UpdateAndRender()
}

// filterRegex will return a regular expression object to match the user's filter input.
func filterRegex() *regexp.Regexp {
	if controllers.Filter == nil || controllers.Filter.view == nil {
		return nil
	}
	filterString := strings.TrimSpace(controllers.Filter.view.Buffer())
	if len(filterString) == 0 {
		return nil
	}

	regex, err := regexp.Compile(filterString)
	if err != nil {
		return nil
	}

	return regex
}

// OnLayoutChange is called by the UI framework to inform the view-model of the new screen dimensions
func (c *FileTree) OnLayoutChange(resized bool) error {
	_ = c.Update()
	if resized {
		return c.Render()
	}
	return nil
}

// Update refreshes the state objects for future rendering.
func (c *FileTree) Update() error {
	var width, height int

	if c.view != nil {
		width, height = c.view.Size()
	} else {
		// before the TUI is setup there may not be a controller to reference. Use the entire screen as reference.
		width, height = c.gui.Size()
	}
	// height should account for the header
	return c.vm.Update(filterRegex(), width, height-1)
}

// Render flushes the state objects (file tree) to the pane.
func (c *FileTree) Render() error {
	title := "Current Layer Contents"
	if controllers.Layer.CompareMode == CompareAll {
		title = "Aggregated Layer Contents"
	}

	// indicate when selected
	if c.gui.CurrentView() == c.view {
		title = "● " + title
	}

	c.gui.Update(func(g *gocui.Gui) error {
		// update the header
		c.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		if c.vm.ShowAttributes {
			headerStr += fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		}

		_, _ = fmt.Fprintln(c.header, format.Header(vtclean.Clean(headerStr, false)))

		// update the contents
		c.view.Clear()
		err := c.vm.Render()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(c.view, c.vm.Buffer.String())

		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (c *FileTree) KeyHelp() string {
	var help string
	for _, binding := range c.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}
