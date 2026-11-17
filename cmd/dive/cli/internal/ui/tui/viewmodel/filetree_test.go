package viewmodel

import (
	"archive/tar"
	"context"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/filetree"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/internal/imagefixture"
)

var updateSnapshot = flag.Bool("update", false, "update any test snapshots")

func TestUpdateSnapshotDisabled(t *testing.T) {
	require.False(t, *updateSnapshot, "update snapshot flag should be disabled")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func testCaseDataFilePath(name string) string {
	return filepath.Join("testdata", name+".txt")
}

func helperLoadBytes(t *testing.T) []byte {
	t.Helper()
	path := testCaseDataFilePath(t.Name())
	theBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to load test data ('%s'): %+v", t.Name(), err)
	}
	return theBytes
}

func helperCaptureBytes(t *testing.T, data []byte) {
	// TODO: switch to https://github.com/gkampitakis/go-snaps
	t.Helper()
	if *updateSnapshot {
		t.Fatalf("cannot capture data in test mode: %s", t.Name())
	}

	path := testCaseDataFilePath(t.Name())
	err := os.WriteFile(path, data, 0644)

	if err != nil {
		t.Fatalf("unable to save test data ('%s'): %+v", t.Name(), err)
	}
}

func initializeTestViewModel(t *testing.T) *FileTreeViewModel {
	t.Helper()

	img := imagefixture.Load(t, "test-docker-image")
	result, err := dive.Analyze(context.Background(), "test-docker-image", img)
	require.NoError(t, err, "unable to analyze test data")

	vm, err := NewFileTreeViewModel(tui.Config{
		Analysis:    result,
		Preferences: tui.DefaultPreferences(),
	}, 0)

	require.NoError(t, err, "unable to create viewmodel")
	return vm
}

func runTestCase(t *testing.T, vm *FileTreeViewModel, width, height int, filterRegex *regexp.Regexp) {
	t.Helper()
	err := vm.Update(filterRegex, width, height)
	if err != nil {
		t.Errorf("failed to update viewmodel: %v", err)
	}

	err = vm.Render()
	if err != nil {
		t.Errorf("failed to render viewmodel: %v", err)
	}

	actualBytes := vm.Buffer.Bytes()
	path := testCaseDataFilePath(t.Name())
	if !fileExists(path) {
		if *updateSnapshot {
			helperCaptureBytes(t, actualBytes)
		} else {
			t.Fatalf("missing test data: %s", path)
		}
	}
	expectedBytes := helperLoadBytes(t)
	if d := cmp.Diff(string(expectedBytes), string(actualBytes)); d != "" {
		t.Errorf("bytes mismatch (-want +got):\n%s", d)
	}
}

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}

func TestFileTreeGoCase(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowAttributes = true

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeNoAttributes(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowAttributes = false

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeRestrictedHeight(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 20
	vm.Setup(0, height)
	vm.ShowAttributes = false

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeDirCollapse(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	assertPath(t, vm, "/bin", "before toggle of bin")

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")
	assertPath(t, vm, "/bin", "after toggle of bin")

	moved := vm.CursorDown() // select /dev
	require.True(t, moved, "unable to cursor down")
	assertPath(t, vm, "/dev", "down to dev")

	moved = vm.CursorDown() // select /etc
	require.True(t, moved, "unable to cursor down")
	assertPath(t, vm, "/etc", "down to etc")

	// collapse /etc
	err = vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /etc")
	assertPath(t, vm, "/etc", "after toggle of etc")

	runTestCase(t, vm, width, height, nil)
}

func assertPath(t *testing.T, vm *FileTreeViewModel, expected string, msg string) {
	t.Helper()
	n := vm.CurrentNode()
	require.NotNil(t, n, "unable to get current node")
	assert.Equal(t, expected, n.Path(), msg)
}

func TestFileTreeCursorLeft(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	require.NoError(t, vm.Update(nil, width, height))

	// /bin -> its first child
	require.True(t, vm.CursorDown())
	n := vm.CurrentNode()
	require.NotNil(t, n)
	assert.Equal(t, "/bin", n.Parent.Path())

	require.NoError(t, vm.CursorLeft())
	assertPath(t, vm, "/bin", "left to parent")
	assert.Equal(t, 0, vm.TreeIndex)

	// a top level entry has no parent row to move to, so the cursor stays put
	require.NoError(t, vm.CursorLeft())
	assertPath(t, vm, "/bin", "left at top level")
}

func TestFileTreeDirCollapseAll(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	err := vm.ToggleCollapseAll()
	checkError(t, err, "unable to collapse all dir")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeSelectLayer(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	// select the next layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 1)
	if err != nil {
		t.Errorf("unable to SetTreeByLayer: %v", err)
	}
	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeSortOrderKeptAcrossLayers(t *testing.T) {
	vm := initializeTestViewModel(t)

	require.NoError(t, vm.ToggleSortOrder())
	require.Equal(t, filetree.SortOrder(filetree.BySizeDesc), vm.ModelTree.SortOrder)

	require.NoError(t, vm.SetTreeByLayer(0, 0, 1, 1))
	assert.Equal(t, filetree.SortOrder(filetree.BySizeDesc), vm.ModelTree.SortOrder)

	// toggling back and revisiting a layer must not flip it again
	require.NoError(t, vm.ToggleSortOrder())
	require.NoError(t, vm.SetTreeByLayer(0, 0, 0, 0))
	assert.Equal(t, filetree.SortOrder(filetree.ByName), vm.ModelTree.SortOrder)
}

func TestFileShowAggregateChanges(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	// select the next layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 13)
	checkError(t, err, "unable to SetTreeByLayer")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreePageDown(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 10
	vm.Setup(0, height)
	vm.ShowAttributes = true
	err := vm.Update(nil, width, height)
	checkError(t, err, "unable to update")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreePageUp(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 10
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// these operations have a render step for intermediate results, which require at least one update to be done first
	err := vm.Update(nil, width, height)
	checkError(t, err, "unable to update")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageDown()
	checkError(t, err, "unable to page down")

	err = vm.PageUp()
	checkError(t, err, "unable to page up")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeDirCursorRight(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	moved := vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	moved = vm.CursorDown()
	if !moved {
		t.Error("unable to cursor down")
	}

	// collapse /etc
	err = vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /etc")

	// expand /etc
	err = vm.CursorRight()
	checkError(t, err, "unable to cursor right")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeFilterTree(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)
	vm.ShowAttributes = true

	regex, err := regexp.Compile("network")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}

func TestFileTreeNavigateWithActiveFilter(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)

	// /etc doesn't match but stays visible since a descendant does
	require.NoError(t, vm.Update(regexp.MustCompile("network"), width, height))
	node := vm.CurrentNode()
	require.NotNil(t, node, "expected a selectable node under an active filter")
	assert.Equal(t, "/etc", node.Path())

	// the selection must line up with the rendered rows
	require.True(t, vm.CursorDown())
	assertPath(t, vm, "/etc/network", "down under filter")

	require.NoError(t, vm.CursorLeft())
	assertPath(t, vm, "/etc", "left under filter")

	require.NoError(t, vm.ToggleCollapse())
	assert.True(t, node.Data.ViewInfo.Collapsed, "expected /etc to collapse under the filter")

	require.NoError(t, vm.CursorRight())
	assert.False(t, node.Data.ViewInfo.Collapsed, "expected /etc to expand under the filter")
	assertPath(t, vm, "/etc/network", "right under filter")
}

func TestFileTreeNavigateWithEmptyFilterResult(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)

	require.NoError(t, vm.Update(regexp.MustCompile("does-not-exist"), width, height))
	assert.Nil(t, vm.CurrentNode())
	require.NoError(t, vm.CursorLeft())
	require.NoError(t, vm.CursorRight())
	require.NoError(t, vm.ToggleCollapse())
}

func TestFileTreeCursorDownStopsAtLastRow(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 1000
	vm.Setup(0, height)

	require.NoError(t, vm.ToggleCollapseAll())
	require.NoError(t, vm.Update(nil, width, height))
	rows := len(strings.Split(strings.TrimSuffix(vm.ViewTree.String(false), "\n"), "\n"))
	for vm.CursorDown() {
	}
	assert.Equal(t, rows-1, vm.TreeIndex)
	assert.NotNil(t, vm.CurrentNode())

	// every collapsed dir is unmodified in the first layer, so nothing is left to select
	vm.ToggleShowDiffType(filetree.Unmodified)
	require.NoError(t, vm.Update(nil, width, height))
	vm.ResetCursor()
	assert.False(t, vm.CursorDown(), "cursor moved past the end of an empty tree")
}

func TestFileTreeCursorClampedOnUpdate(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 20
	vm.Setup(0, height)
	require.NoError(t, vm.Update(nil, width, height))

	assertCursorInPane := func(msg string) {
		t.Helper()
		assert.GreaterOrEqual(t, vm.bufferIndex, 0, msg)
		assert.LessOrEqual(t, vm.bufferIndex, vm.height(), msg)
		assert.Equal(t, vm.TreeIndex, vm.bufferIndexLowerBound+vm.bufferIndex, msg)
		assert.NotNil(t, vm.CurrentNode(), msg)
	}

	for range 200 {
		require.True(t, vm.CursorDown())
	}

	// shrinking the pane keeps the cursor on the same node, inside the pane
	before := vm.CurrentNode()
	require.NoError(t, vm.Update(nil, width, 10))
	assertCursorInPane("after resize")
	assert.Equal(t, before, vm.CurrentNode())

	// filtering down to a handful of rows moves the cursor onto the last of them and shows them all
	require.NoError(t, vm.Update(regexp.MustCompile("network"), width, 10))
	assertCursorInPane("after filter")
	assert.Equal(t, vm.ModelTree.VisibleSize(), vm.TreeIndex)
	assert.Equal(t, 0, vm.bufferIndexLowerBound)

	// hiding everything leaves the cursor at the top
	vm.ToggleShowDiffType(filetree.Unmodified)
	require.NoError(t, vm.Update(nil, width, 10))
	assert.Equal(t, 0, vm.TreeIndex)
	assert.Equal(t, 0, vm.bufferIndex)
	assert.Equal(t, 0, vm.bufferIndexLowerBound)
}

func TestFileTreeHideAddedRemovedModified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to SetTreeByLayer: %v", err)
	}

	// hide added files
	vm.ToggleShowDiffType(filetree.Added)

	// hide modified files
	vm.ToggleShowDiffType(filetree.Modified)

	// hide removed files
	vm.ToggleShowDiffType(filetree.Removed)

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideUnmodified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to SetTreeByLayer: %v", err)
	}

	// hide unmodified files
	vm.ToggleShowDiffType(filetree.Unmodified)

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideTypeWithFilter(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse()
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to SetTreeByLayer: %v", err)
	}

	// hide added files
	vm.ToggleShowDiffType(filetree.Added)

	regex, err := regexp.Compile("saved")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}

// regression for #306: kaniko built and Windows images flooded the TUI with "path error at layer index ...: unable
// to remove/add ..." and refused to start without --ignore-errors. The layers here whiteout paths no lower layer
// has (as in the Windows images), repeat the same whiteout in later layers (as kaniko does for apt's auxfiles) and
// carry a character device.
func TestFileTreeWhiteoutsOfMissingPaths(t *testing.T) {
	dir := func(name string) imagefixture.Entry { return imagefixture.Entry{Name: name, Type: tar.TypeDir} }
	file := func(name string) imagefixture.Entry { return imagefixture.Entry{Name: name, Content: name} }
	whiteout := func(name string) imagefixture.Entry { return imagefixture.Entry{Name: name} }
	aptLists := []imagefixture.Entry{dir("var/"), dir("var/lib/"), dir("var/lib/apt/"), dir("var/lib/apt/lists/")}
	scm := []imagefixture.Entry{dir("Files/"), dir("Files/Windows/"), dir("Files/Windows/System32/"), dir("Files/Windows/System32/LogFiles/"), dir("Files/Windows/System32/LogFiles/Scm/")}

	layers := [][]imagefixture.Entry{
		append(append(slices.Clone(aptLists),
			dir("var/lib/apt/lists/auxfiles/"),
			file("var/lib/apt/lists/auxfiles/a"),
			file("var/lib/apt/lists/lock"),
			dir("dev/"),
			imagefixture.Entry{Name: "dev/null", Type: tar.TypeChar, Mode: 0o666},
			dir("Hives/"),
			file("Hives/DefaultUser_Delta"),
		), scm...),
		append(append(slices.Clone(aptLists),
			whiteout("var/lib/apt/lists/.wh.auxfiles"),
		), append(slices.Clone(scm),
			whiteout("Files/Windows/System32/LogFiles/Scm/.wh.scm.evm"),
			file("Files/Windows/System32/LogFiles/Scm/new.evm"),
		)...),
		append(slices.Clone(aptLists), whiteout("var/lib/apt/lists/.wh.auxfiles"), file("var/lib/apt/lists/two")),
		append(slices.Clone(aptLists), whiteout("var/lib/apt/lists/.wh.auxfiles"), file("var/lib/apt/lists/three")),
	}
	img := imagefixture.LoadSpec(t, imagefixture.Spec{Format: imagefixture.OCISave, Compression: imagefixture.Gzip, Layers: layers})
	analysis, err := dive.Analyze(context.Background(), "spec", img)
	require.NoError(t, err)

	vm, err := NewFileTreeViewModel(tui.Config{Analysis: analysis, Preferences: tui.DefaultPreferences()}, 0)
	require.NoError(t, err)

	// every layer selection the TUI can make builds a tree
	for _, mode := range []LayerCompareMode{CompareSingleLayer, CompareAllLayers} {
		for idx := range layers {
			state := LayerSetState{LayerIndex: idx, CompareMode: mode}
			require.NoError(t, vm.SetTreeByLayer(state.GetCompareIndexes()), "layer %d mode %d", idx, mode)
		}
	}

	marks := func(t *testing.T, bottomStop, topStart, topStop int, want map[string]filetree.DiffType, absent ...string) {
		t.Helper()
		require.NoError(t, vm.SetTreeByLayer(0, bottomStop, topStart, topStop))
		for p, d := range want {
			n, err := vm.ModelTree.GetNode(p)
			if assert.NoError(t, err, p) {
				assert.Equal(t, d, n.Data.DiffType, p)
			}
		}
		for _, p := range absent {
			_, err := vm.ModelTree.GetNode(p)
			assert.Error(t, err, p)
		}
	}

	// the valid whiteout removes auxfiles, the one with no target is skipped and its sibling still counts
	marks(t, 0, 1, 1, map[string]filetree.DiffType{
		"/var/lib/apt/lists/auxfiles":                  filetree.Removed,
		"/var/lib/apt/lists/auxfiles/a":                filetree.Removed,
		"/var/lib/apt/lists/lock":                      filetree.Unmodified,
		"/Files/Windows/System32/LogFiles/Scm/new.evm": filetree.Added,
		"/dev/null":                filetree.Unmodified,
		"/Hives/DefaultUser_Delta": filetree.Unmodified,
	}, "/Files/Windows/System32/LogFiles/Scm/scm.evm")

	// the repeated whiteout has nothing left to remove
	marks(t, 1, 2, 2, map[string]filetree.DiffType{
		"/var/lib/apt/lists/two":  filetree.Added,
		"/var/lib/apt/lists/lock": filetree.Unmodified,
	}, "/var/lib/apt/lists/auxfiles")

	// aggregated: everything after the base layer at once
	marks(t, 0, 1, 3, map[string]filetree.DiffType{
		"/var/lib/apt/lists/auxfiles": filetree.Removed,
		"/var/lib/apt/lists/two":      filetree.Added,
		"/var/lib/apt/lists/three":    filetree.Added,
	})
}
