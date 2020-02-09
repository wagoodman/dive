package view

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/runtime/ui/format"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
	"github.com/wagoodman/dive/utils"
	"regexp"
)

type ViewOptionChangeListener func() error

// FileTree holds the UI objects and data models for populating the right pane. Specifically the pane that
// shows selected layer or aggregate file ASCII tree.
type FileTree struct {
	name   string
	gui    *gocui.Gui
	view   *gocui.View
	header *gocui.View
	vm     *viewmodel.FileTree
	title  string

	filterRegex         *regexp.Regexp
	listeners           []ViewOptionChangeListener
	helpKeys            []*key.Binding
	requestedWidthRatio float64
}

// newFileTreeView creates a new view object attached the the global [gocui] screen object.
func newFileTreeView(gui *gocui.Gui, tree *filetree.FileTree, refTrees []*filetree.FileTree, cache filetree.Comparer) (controller *FileTree, err error) {
	controller = new(FileTree)
	controller.listeners = make([]ViewOptionChangeListener, 0)

	// populate main fields
	controller.name = "filetree"
	controller.gui = gui
	controller.vm, err = viewmodel.NewFileTreeViewModel(tree, refTrees, cache)
	if err != nil {
		return nil, err
	}

	requestedWidthRatio := viper.GetFloat64("filetree.pane-width")
	if requestedWidthRatio >= 1 || requestedWidthRatio <= 0 {
		logrus.Errorf("invalid config value: 'filetree.pane-width' should be 0 < value < 1, given '%v'", requestedWidthRatio)
		requestedWidthRatio = 0.5
	}
	controller.requestedWidthRatio = requestedWidthRatio

	return controller, err
}

func (v *FileTree) AddViewOptionChangeListener(listener ...ViewOptionChangeListener) {
	v.listeners = append(v.listeners, listener...)
}

func (v *FileTree) SetTitle(title string) {
	v.title = title
}

func (v *FileTree) SetFilterRegex(filterRegex *regexp.Regexp) {
	v.filterRegex = filterRegex
}

func (v *FileTree) Name() string {
	return v.name
}

// Setup initializes the UI concerns within the context of a global [gocui] view object.
func (v *FileTree) Setup(view *gocui.View, header *gocui.View) error {
	logrus.Tracef("view.Setup() %s", v.Name())

	// set controller options
	v.view = view
	v.view.Editable = false
	v.view.Wrap = false
	v.view.Frame = false

	v.header = header
	v.header.Editable = false
	v.header.Wrap = false
	v.header.Frame = false

	var infos = []key.BindingInfo{
		{
			ConfigKeys: []string{"keybinding.toggle-collapse-dir"},
			OnAction:   v.toggleCollapse,
			Display:    "Collapse dir",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-collapse-all-dir"},
			OnAction:   v.toggleCollapseAll,
			Display:    "Collapse all dir",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-added-files"},
			OnAction:   func() error { return v.toggleShowDiffType(filetree.Added) },
			IsSelected: func() bool { return !v.vm.HiddenDiffTypes[filetree.Added] },
			Display:    "Added",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-removed-files"},
			OnAction:   func() error { return v.toggleShowDiffType(filetree.Removed) },
			IsSelected: func() bool { return !v.vm.HiddenDiffTypes[filetree.Removed] },
			Display:    "Removed",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-modified-files"},
			OnAction:   func() error { return v.toggleShowDiffType(filetree.Modified) },
			IsSelected: func() bool { return !v.vm.HiddenDiffTypes[filetree.Modified] },
			Display:    "Modified",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-unchanged-files", "keybinding.toggle-unmodified-files"},
			OnAction:   func() error { return v.toggleShowDiffType(filetree.Unmodified) },
			IsSelected: func() bool { return !v.vm.HiddenDiffTypes[filetree.Unmodified] },
			Display:    "Unmodified",
		},
		{
			ConfigKeys: []string{"keybinding.toggle-filetree-attributes"},
			OnAction:   v.toggleAttributes,
			IsSelected: func() bool { return v.vm.ShowAttributes },
			Display:    "Attributes",
		},
		{
			ConfigKeys: []string{"keybinding.page-up"},
			OnAction:   v.PageUp,
		},
		{
			ConfigKeys: []string{"keybinding.page-down"},
			OnAction:   v.PageDown,
		},
		{
			Key:      gocui.KeyArrowDown,
			Modifier: gocui.ModNone,
			OnAction: v.CursorDown,
		},
		{
			Key:      gocui.KeyArrowUp,
			Modifier: gocui.ModNone,
			OnAction: v.CursorUp,
		},
		{
			Key:      gocui.KeyArrowLeft,
			Modifier: gocui.ModNone,
			OnAction: v.CursorLeft,
		},
		{
			Key:      gocui.KeyArrowRight,
			Modifier: gocui.ModNone,
			OnAction: v.CursorRight,
		},
	}

	helpKeys, err := key.GenerateBindings(v.gui, v.name, infos)
	if err != nil {
		return err
	}
	v.helpKeys = helpKeys

	_, height := v.view.Size()
	v.vm.Setup(0, height)
	_ = v.Update()
	_ = v.Render()

	return nil
}

// IsVisible indicates if the file tree view pane is currently initialized
func (v *FileTree) IsVisible() bool {
	return v != nil
}

// ResetCursor moves the cursor back to the top of the buffer and translates to the top of the buffer.
func (v *FileTree) resetCursor() {
	_ = v.view.SetCursor(0, 0)
	v.vm.ResetCursor()
}

// SetTreeByLayer populates the view model by stacking the indicated image layer file trees.
func (v *FileTree) SetTree(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop int) error {
	err := v.vm.SetTreeByLayer(bottomTreeStart, bottomTreeStop, topTreeStart, topTreeStop)
	if err != nil {
		return err
	}

	_ = v.Update()
	return v.Render()
}

// CursorDown moves the cursor down and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (v *FileTree) CursorDown() error {
	if v.vm.CursorDown() {
		return v.Render()
	}
	return nil
}

// CursorUp moves the cursor up and renders the view.
// Note: we cannot use the gocui buffer since any state change requires writing the entire tree to the buffer.
// Instead we are keeping an upper and lower bounds of the tree string to render and only flushing
// this range into the view buffer. This is much faster when tree sizes are large.
func (v *FileTree) CursorUp() error {
	if v.vm.CursorUp() {
		return v.Render()
	}
	return nil
}

// CursorLeft moves the cursor up until we reach the Parent Node or top of the tree
func (v *FileTree) CursorLeft() error {
	err := v.vm.CursorLeft(v.filterRegex)
	if err != nil {
		return err
	}
	_ = v.Update()
	return v.Render()
}

// CursorRight descends into directory expanding it if needed
func (v *FileTree) CursorRight() error {
	err := v.vm.CursorRight(v.filterRegex)
	if err != nil {
		return err
	}
	_ = v.Update()
	return v.Render()
}

// PageDown moves to next page putting the cursor on top
func (v *FileTree) PageDown() error {
	err := v.vm.PageDown()
	if err != nil {
		return err
	}
	return v.Render()
}

// PageUp moves to previous page putting the cursor on top
func (v *FileTree) PageUp() error {
	err := v.vm.PageUp()
	if err != nil {
		return err
	}
	return v.Render()
}

// getAbsPositionNode determines the selected screen cursor's location in the file tree, returning the selected FileNode.
// func (controller *FileTree) getAbsPositionNode() (node *filetree.FileNode) {
// 	return controller.vm.getAbsPositionNode(filterRegex())
// }

// ToggleCollapse will collapse/expand the selected FileNode.
func (v *FileTree) toggleCollapse() error {
	err := v.vm.ToggleCollapse(v.filterRegex)
	if err != nil {
		return err
	}
	_ = v.Update()
	return v.Render()
}

// ToggleCollapseAll will collapse/expand the all directories.
func (v *FileTree) toggleCollapseAll() error {
	err := v.vm.ToggleCollapseAll()
	if err != nil {
		return err
	}
	if v.vm.CollapseAll {
		v.resetCursor()
	}
	_ = v.Update()
	return v.Render()
}

func (v *FileTree) notifyOnViewOptionChangeListeners() error {
	for _, listener := range v.listeners {
		err := listener()
		if err != nil {
			logrus.Errorf("notifyOnViewOptionChangeListeners error: %+v", err)
			return err
		}
	}
	return nil
}

// ToggleAttributes will show/hide file attributes
func (v *FileTree) toggleAttributes() error {
	err := v.vm.ToggleAttributes()
	if err != nil {
		return err
	}

	err = v.Update()
	if err != nil {
		return err
	}
	err = v.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/view)
	return v.notifyOnViewOptionChangeListeners()
}

// ToggleShowDiffType will show/hide the selected DiffType in the filetree pane.
func (v *FileTree) toggleShowDiffType(diffType filetree.DiffType) error {
	v.vm.ToggleShowDiffType(diffType)

	err := v.Update()
	if err != nil {
		return err
	}
	err = v.Render()
	if err != nil {
		return err
	}

	// we need to render the changes to the status pane as well (not just this contoller/view)
	return v.notifyOnViewOptionChangeListeners()
}

// OnLayoutChange is called by the UI framework to inform the view-model of the new screen dimensions
func (v *FileTree) OnLayoutChange() error {
	err := v.Update()
	if err != nil {
		return err
	}
	return v.Render()
}

// Update refreshes the state objects for future rendering.
func (v *FileTree) Update() error {
	var width, height int

	if v.view != nil {
		width, height = v.view.Size()
	} else {
		// before the TUI is setup there may not be a controller to reference. Use the entire screen as reference.
		width, height = v.gui.Size()
	}
	// height should account for the header
	return v.vm.Update(v.filterRegex, width, height-1)
}

// Render flushes the state objects (file tree) to the pane.
func (v *FileTree) Render() error {
	logrus.Tracef("view.Render() %s", v.Name())

	title := v.title
	isSelected := v.gui.CurrentView() == v.view

	v.gui.Update(func(g *gocui.Gui) error {
		// update the header
		v.header.Clear()
		width, _ := g.Size()
		headerStr := format.RenderHeader(title, width, isSelected)
		if v.vm.ShowAttributes {
			headerStr += fmt.Sprintf(filetree.AttributeFormat+" %s", "P", "ermission", "UID:GID", "Size", "Filetree")
		}
		_, _ = fmt.Fprintln(v.header, headerStr)

		// update the contents
		v.view.Clear()
		err := v.vm.Render()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(v.view, v.vm.Buffer.String())

		return err
	})
	return nil
}

// KeyHelp indicates all the possible actions a user can take while the current pane is selected.
func (v *FileTree) KeyHelp() string {
	var help string
	for _, binding := range v.helpKeys {
		help += binding.RenderKeyHelp()
	}
	return help
}

func (v *FileTree) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	logrus.Tracef("view.Layout(minX: %d, minY: %d, maxX: %d, maxY: %d) %s", minX, minY, maxX, maxY, v.Name())
	attributeRowSize := 0

	// make the layout responsive to the available realestate. Make more room for the main content by hiding auxillary
	// content when there is not enough room
	if maxX-minX < 60 {
		v.vm.ConstrainLayout()
	} else {
		v.vm.ExpandLayout()
	}

	if v.vm.ShowAttributes {
		attributeRowSize = 1
	}

	// header + attribute header
	headerSize := 1 + attributeRowSize
	// note: maxY needs to account for the (invisible) border, thus a +1
	header, headerErr := g.SetView(v.Name()+"header", minX, minY, maxX, minY+headerSize+1)
	// we are going to overlap the view over the (invisible) border (so minY will be one less than expected).
	// additionally, maxY will be bumped by one to include the border
	view, viewErr := g.SetView(v.Name(), minX, minY+headerSize, maxX, maxY+1)
	if utils.IsNewView(viewErr, headerErr) {
		err := v.Setup(view, header)
		if err != nil {
			logrus.Error("unable to setup tree controller", err)
			return err
		}
	}
	return nil
}

func (v *FileTree) RequestedSize(available int) *int {
	//var requestedWidth = int(float64(available) * (1.0 - v.requestedWidthRatio))
	//return &requestedWidth
	return nil
}
