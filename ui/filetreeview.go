package ui

import (
	"fmt"
	"github.com/lunixbochs/vtclean"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/keybinding"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/dive/filetree"
)

const (
	CompareLayer CompareType = iota
	CompareAll
)

type CompareType int

// FileTreeView holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
type FileTreeView struct {
	Name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	vm     *FileTreeViewModel

	keybindingToggleCollapse    []keybinding.Key
	keybindingToggleCollapseAll []keybinding.Key
	keybindingToggleAttributes    []keybinding.Key
	keybindingToggleAdded       []keybinding.Key
	keybindingToggleRemoved     []keybinding.Key
	keybindingToggleModified    []keybinding.Key
	keybindingToggleUnchanged   []keybinding.Key
	keybindingPageDown          []keybinding.Key
	keybindingPageUp            []keybinding.Key
}

// NewFileTreeView creates a new view object attached the the global [gocui] screen object.
func NewFileTreeView(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree, cache filetree.TreeCache) (treeView *FileTreeView) {
	treeView = new(FileTreeView)

	// populate main fields
	treeView.Name = name
	treeView.gui = gui
	treeView.vm = NewFileTreeViewModel(tree, refTrees, cache)

	var err error
	treeView.keybindingToggleCollapse, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-collapse-dir"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleCollapseAll, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-collapse-all-dir"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleAttributes, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-filetree-attributes"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleAdded, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-added-files"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleRemoved, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-removed-files"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleModified, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-modified-files"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingToggleUnchanged, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-unchanged-files"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingPageUp, err = keybinding.ParseAll(viper.GetString("keybinding.page-up"))
	if err != nil {
		logrus.Error(err)
	}

	treeView.keybindingPageDown, err = keybinding.ParseAll(viper.GetString("keybinding.page-down"))
	if err != nil {
		logrus.Error(err)
	}

	return treeView
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (view *FileTreeView) Setup(v *gocui.View, header *gocui.View) error {

	// set view options
	view.view = v
	view.view.Editable = false
	view.view.Wrap = false
	view.view.Frame = false

	view.header = header
	view.header.Editable = false
	view.header.Wrap = false
	view.header.Frame = false

	// set keybindings
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorDown() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorUp() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowLeft, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorLeft() }); err != nil {
		return err
	}
	if err := view.gui.SetKeybinding(view.Name, gocui.KeyArrowRight, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return view.CursorRight() }); err != nil {
		return err
	}

	for _, key := range view.keybindingPageUp {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.PageUp() }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingPageDown {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.PageDown() }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleCollapse {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleCollapse() }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleCollapseAll {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleCollapseAll() }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleAttributes {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleAttributes() }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleAdded {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Added) }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleRemoved {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Removed) }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleModified {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Changed) }); err != nil {
			return err
		}
	}
	for _, key := range view.keybindingToggleUnchanged {
		if err := view.gui.SetKeybinding(view.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return view.toggleShowDiffType(filetree.Unchanged) }); err != nil {
			return err
		}
	}

	_, height := view.view.Size()
	view.vm.Setup(0, height)
	view.Update()
	view.Render()

	return nil
}

// IsVisible indicates if the file tree view pane is currently initialized
func (view *FileTreeView) IsVisible() bool {
	if view == nil {
		return false
	}
	return true
}

// resetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (view *FileTreeView) resetCursor() {
	view.view.SetCursor(0, 0)
	view.vm.resetCursor()
}

// setTreeByLayer populates the view model by stacking the indicated image layer file trees.
func (view *FileTreeView) setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	err := view.vm.setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	if err != nil {
		return err
	}
	// view.resetCursor()

	view.Update()
	return view.Render()
}

// CursorDown moves the cursor down and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (view *FileTreeView) CursorDown() error {
	if view.vm.CursorDown() {
		return view.Render()
	}
	return nil
}

// CursorUp moves the cursor up and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (view *FileTreeView) CursorUp() error {
	if view.vm.CursorUp() {
		return view.Render()
	}
	return nil
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (view *FileTreeView) CursorLeft() error {
	err := view.vm.CursorLeft(filterRegex())
	if err != nil {
		return err
	}
	view.Update()
	return view.Render()
}

// CursorRight descends into directory expanding it if needed
func (view *FileTreeView) CursorRight() error {
	err := view.vm.CursorRight(filterRegex())
	if err != nil {
		return err
	}
	view.Update()
	return view.Render()
}

// PageDown moves to next page putting the cursor on top
func (view *FileTreeView) PageDown() error {
	err := view.vm.PageDown()
	if err != nil {
		return err
	}
	return view.Render()
}

// PageUp moves to previous page putting the cursor on top
func (view *FileTreeView) PageUp() error {
	err := view.vm.PageUp()
	if err != nil {
		return err
	}
	return view.Render()
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
func (view *FileTreeView) getAbsPositionNode() (node *filetree.FileNode) {
	return view.vm.getAbsPositionNode(filterRegex())
}

// toggleCollapse will collapse/expand the selected FileNode.
func (view *FileTreeView) toggleCollapse() error {
	err := view.vm.toggleCollapse(filterRegex())
	if err != nil {
		return err
	}
	view.Update()
	return view.Render()
}

// toggleCollapseAll will collapse/expand the all directories.
func (view *FileTreeView) toggleCollapseAll() error {
	err := view.vm.toggleCollapseAll(filterRegex())
	if err != nil {
		return err
	}
	view.Update()
	return view.Render()
}

// toggleAttributes will show/hide file attributes
func (view *FileTreeView) toggleAttributes() error {
	err := view.vm.toggleAttributes()
	if err != nil {
		return err
	}
	// we need to render the changes to the status pane as well
	Update()
	Render()
	return nil
}

// toggleShowDiffType will show/hide the selected DiffType in the filetree pane.
func (view *FileTreeView) toggleShowDiffType(diffType filetree.DiffType) error {
	view.vm.toggleShowDiffType(diffType)
	// we need to render the changes to the status pane as well
	Update()
	Render()
	return nil
}

// filterRegex will return a regular expression object to match the user's filter input.
func filterRegex() *regexp.Regexp {
	if Views.Filter == nil || Views.Filter.view == nil {
		return nil
	}
	filterString := strings.TrimSpace(Views.Filter.view.Buffer())
	if len(filterString) == 0 {
		return nil
	}

	regex, err := regexp.Compile(filterString)
	if err != nil {
		return nil
	}

	return regex
}

// Update refreshes the state objects for future rendering.
func (view *FileTreeView) Update() error {
	var width, height int

	if view.view != nil {
		width, height = view.view.Size()
	} else {
		// before the TUI is setup there may not be a view to reference. Use the entire screen as reference.
		width, height = view.gui.Size()
	}
	// height should account for the header
	return view.vm.Update(filterRegex(), width, height-1)
}

// Render flushes the state objects (file tree) to the pane.
func (view *FileTreeView) Render() error {
	title := "Current Layer Contents"
	if Views.Layer.CompareMode == CompareAll {
		title = "Aggregated Layer Contents"
	}

	// indicate when selected
	if view.gui.CurrentView() == view.view {
		title = "● " + title
	}

	view.gui.Update(func(g *gocui.Gui) error {
		// update the header
		view.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		if view.vm.ShowAttributes {
			headerStr += fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		}

		fmt.Fprintln(view.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update the contents
		view.view.Clear()
		view.vm.Render()
		fmt.Fprint(view.view, view.vm.mainBuf.String())

		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (view *FileTreeView) KeyHelp() string {
	return renderStatusOption(view.keybindingToggleCollapse[0].String(), "Collapse dir", false) +
		renderStatusOption(view.keybindingToggleCollapseAll[0].String(), "Collapse all dir", false) +
		renderStatusOption(view.keybindingToggleAdded[0].String(), "Added", !view.vm.HiddenDiffTypes[filetree.Added]) +
		renderStatusOption(view.keybindingToggleRemoved[0].String(), "Removed", !view.vm.HiddenDiffTypes[filetree.Removed]) +
		renderStatusOption(view.keybindingToggleModified[0].String(), "Modified", !view.vm.HiddenDiffTypes[filetree.Changed]) +
		renderStatusOption(view.keybindingToggleUnchanged[0].String(), "Unmodified", !view.vm.HiddenDiffTypes[filetree.Unchanged]) +
	    renderStatusOption(view.keybindingToggleAttributes[0].String(), "Attributes", view.vm.ShowAttributes)
}
