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

// FileTreeController holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
type FileTreeController struct {
	Name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	vm     *FileTreeViewModel

	keybindingToggleCollapse    []keybinding.Key
	keybindingToggleCollapseAll []keybinding.Key
	keybindingToggleAttributes  []keybinding.Key
	keybindingToggleAdded       []keybinding.Key
	keybindingToggleRemoved     []keybinding.Key
	keybindingToggleModified    []keybinding.Key
	keybindingToggleUnchanged   []keybinding.Key
	keybindingPageDown          []keybinding.Key
	keybindingPageUp            []keybinding.Key
}

// NewFileTreeController creates a new view object attached the the global [gocui] screen object.
func NewFileTreeController(name string, gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree, cache filetree.TreeCache) (controller *FileTreeController) {
	controller = new(FileTreeController)

	// populate main fields
	controller.Name = name
	controller.gui = gui
	controller.vm = NewFileTreeViewModel(tree, refTrees, cache)

	var err error
	controller.keybindingToggleCollapse, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-collapse-dir"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleCollapseAll, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-collapse-all-dir"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleAttributes, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-filetree-attributes"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleAdded, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-added-files"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleRemoved, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-removed-files"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleModified, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-modified-files"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingToggleUnchanged, err = keybinding.ParseAll(viper.GetString("keybinding.toggle-unchanged-files"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingPageUp, err = keybinding.ParseAll(viper.GetString("keybinding.page-up"))
	if err != nil {
		logrus.Error(err)
	}

	controller.keybindingPageDown, err = keybinding.ParseAll(viper.GetString("keybinding.page-down"))
	if err != nil {
		logrus.Error(err)
	}

	return controller
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (controller *FileTreeController) Setup(v *gocui.View, header *gocui.View) error {

	// set controller options
	controller.view = v
	controller.view.Editable = false
	controller.view.Wrap = false
	controller.view.Frame = false

	controller.header = header
	controller.header.Editable = false
	controller.header.Wrap = false
	controller.header.Frame = false

	// set keybindings
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowDown, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorDown() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowUp, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorUp() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowLeft, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorLeft() }); err != nil {
		return err
	}
	if err := controller.gui.SetKeybinding(controller.Name, gocui.KeyArrowRight, gocui.ModNone, func(*gocui.Gui, *gocui.View) error { return controller.CursorRight() }); err != nil {
		return err
	}

	for _, key := range controller.keybindingPageUp {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.PageUp() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingPageDown {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.PageDown() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleCollapse {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleCollapse() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleCollapseAll {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleCollapseAll() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleAttributes {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleAttributes() }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleAdded {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleShowDiffType(filetree.Added) }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleRemoved {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleShowDiffType(filetree.Removed) }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleModified {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleShowDiffType(filetree.Changed) }); err != nil {
			return err
		}
	}
	for _, key := range controller.keybindingToggleUnchanged {
		if err := controller.gui.SetKeybinding(controller.Name, key.Value, key.Modifier, func(*gocui.Gui, *gocui.View) error { return controller.toggleShowDiffType(filetree.Unchanged) }); err != nil {
			return err
		}
	}

	_, height := controller.view.Size()
	controller.vm.Setup(0, height)
	controller.Update()
	controller.Render()

	return nil
}

// IsVisible indicates if the file tree view pane is currently initialized
func (controller *FileTreeController) IsVisible() bool {
	if controller == nil {
		return false
	}
	return true
}

// resetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (controller *FileTreeController) resetCursor() {
	controller.view.SetCursor(0, 0)
	controller.vm.resetCursor()
}

// setTreeByLayer populates the view model by stacking the indicated image layer file trees.
func (controller *FileTreeController) setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	err := controller.vm.setTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	if err != nil {
		return err
	}
	// controller.resetCursor()

	controller.Update()
	return controller.Render()
}

// CursorDown moves the cursor down and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (controller *FileTreeController) CursorDown() error {
	if controller.vm.CursorDown() {
		return controller.Render()
	}
	return nil
}

// CursorUp moves the cursor up and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (controller *FileTreeController) CursorUp() error {
	if controller.vm.CursorUp() {
		return controller.Render()
	}
	return nil
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (controller *FileTreeController) CursorLeft() error {
	err := controller.vm.CursorLeft(filterRegex())
	if err != nil {
		return err
	}
	controller.Update()
	return controller.Render()
}

// CursorRight descends into directory expanding it if needed
func (controller *FileTreeController) CursorRight() error {
	err := controller.vm.CursorRight(filterRegex())
	if err != nil {
		return err
	}
	controller.Update()
	return controller.Render()
}

// PageDown moves to next page putting the cursor on top
func (controller *FileTreeController) PageDown() error {
	err := controller.vm.PageDown()
	if err != nil {
		return err
	}
	return controller.Render()
}

// PageUp moves to previous page putting the cursor on top
func (controller *FileTreeController) PageUp() error {
	err := controller.vm.PageUp()
	if err != nil {
		return err
	}
	return controller.Render()
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
func (controller *FileTreeController) getAbsPositionNode() (node *filetree.FileNode) {
	return controller.vm.getAbsPositionNode(filterRegex())
}

// toggleCollapse will collapse/expand the selected FileNode.
func (controller *FileTreeController) toggleCollapse() error {
	err := controller.vm.toggleCollapse(filterRegex())
	if err != nil {
		return err
	}
	controller.Update()
	return controller.Render()
}

// toggleCollapseAll will collapse/expand the all directories.
func (controller *FileTreeController) toggleCollapseAll() error {
	err := controller.vm.toggleCollapseAll()
	if err != nil {
		return err
	}
	if controller.vm.CollapseAll {
		controller.resetCursor()
	}
	controller.Update()
	return controller.Render()
}

// toggleAttributes will show/hide file attributes
func (controller *FileTreeController) toggleAttributes() error {
	err := controller.vm.toggleAttributes()
	if err != nil {
		return err
	}
	// we need to render the changes to the status pane as well
	Update()
	Render()
	return nil
}

// toggleShowDiffType will show/hide the selected DiffType in the filetree pane.
func (controller *FileTreeController) toggleShowDiffType(diffType filetree.DiffType) error {
	controller.vm.toggleShowDiffType(diffType)
	// we need to render the changes to the status pane as well
	Update()
	Render()
	return nil
}

// filterRegex will return a regular expression object to match the user's filter input.
func filterRegex() *regexp.Regexp {
	if Controllers.Filter == nil || Controllers.Filter.view == nil {
		return nil
	}
	filterString := strings.TrimSpace(Controllers.Filter.view.Buffer())
	if len(filterString) == 0 {
		return nil
	}

	regex, err := regexp.Compile(filterString)
	if err != nil {
		return nil
	}

	return regex
}

// onLayoutChange is called by the UI framework to inform the view-model of the new screen dimensions
func (controller *FileTreeController) onLayoutChange(resized bool) error {
	controller.Update()
	if resized {
		return controller.Render()
	}
	return nil
}

// Update refreshes the state objects for future rendering.
func (controller *FileTreeController) Update() error {
	var width, height int

	if controller.view != nil {
		width, height = controller.view.Size()
	} else {
		// before the TUI is setup there may not be a controller to reference. Use the entire screen as reference.
		width, height = controller.gui.Size()
	}
	// height should account for the header
	return controller.vm.Update(filterRegex(), width, height-1)
}

// Render flushes the state objects (file tree) to the pane.
func (controller *FileTreeController) Render() error {
	title := "Current Layer Contents"
	if Controllers.Layer.CompareMode == CompareAll {
		title = "Aggregated Layer Contents"
	}

	// indicate when selected
	if controller.gui.CurrentView() == controller.view {
		title = "● " + title
	}

	controller.gui.Update(func(g *gocui.Gui) error {
		// update the header
		controller.header.Clear()
		width, _ := g.Size()
		headerStr := fmt.Sprintf("[%s]%s\n", title, strings.Repeat("─", width*2))
		if controller.vm.ShowAttributes {
			headerStr += fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		}

		fmt.Fprintln(controller.header, Formatting.Header(vtclean.Clean(headerStr, false)))

		// update the contents
		controller.view.Clear()
		controller.vm.Render()
		fmt.Fprint(controller.view, controller.vm.mainBuf.String())

		return nil
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (controller *FileTreeController) KeyHelp() string {
	return renderStatusOption(controller.keybindingToggleCollapse[0].String(), "Collapse dir", false) +
		renderStatusOption(controller.keybindingToggleCollapseAll[0].String(), "Collapse all dir", false) +
		renderStatusOption(controller.keybindingToggleAdded[0].String(), "Added", !controller.vm.HiddenDiffTypes[filetree.Added]) +
		renderStatusOption(controller.keybindingToggleRemoved[0].String(), "Removed", !controller.vm.HiddenDiffTypes[filetree.Removed]) +
		renderStatusOption(controller.keybindingToggleModified[0].String(), "Modified", !controller.vm.HiddenDiffTypes[filetree.Changed]) +
		renderStatusOption(controller.keybindingToggleUnchanged[0].String(), "Unmodified", !controller.vm.HiddenDiffTypes[filetree.Unchanged]) +
		renderStatusOption(controller.keybindingToggleAttributes[0].String(), "Attributes", controller.vm.ShowAttributes)
}
