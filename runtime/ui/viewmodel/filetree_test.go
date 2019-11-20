package viewmodel

import (
	"bytes"
	"github.com/wagoodman/dive/dive/image/docker"
	"github.com/wagoodman/dive/runtime/ui/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/wagoodman/dive/dive/filetree"
)

const allowTestDataCapture = false

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
	path := testCaseDataFilePath(t.Name())
	theBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to load test data ('%s'): %+v", t.Name(), err)
	}
	return theBytes
}

func helperCaptureBytes(t *testing.T, data []byte) {
	if !allowTestDataCapture {
		t.Fatalf("cannot capture data in test mode: %s", t.Name())
	}

	path := testCaseDataFilePath(t.Name())
	err := ioutil.WriteFile(path, data, 0644)

	if err != nil {
		t.Fatalf("unable to save test data ('%s'): %+v", t.Name(), err)
	}
}

func helperCheckDiff(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf(dmp.DiffPrettyText(diffs))
		t.Errorf("%s: bytes mismatch", t.Name())
	}
}

func assertTestData(t *testing.T, actualBytes []byte) {
	path := testCaseDataFilePath(t.Name())
	if !fileExists(path) {
		if allowTestDataCapture {
			helperCaptureBytes(t, actualBytes)
		} else {
			t.Fatalf("missing test data: %s", path)
		}
	}
	expectedBytes := helperLoadBytes(t)
	helperCheckDiff(t, expectedBytes, actualBytes)
}

func initializeTestViewModel(t *testing.T) *FileTree {
	result := docker.TestAnalysisFromArchive(t, "../../../.data/test-docker-image.tar")

	cache := filetree.NewComparer(result.RefTrees)
	errors := cache.BuildCache()
	if len(errors) > 0 {
		t.Fatalf("%s: unable to build cache: %d errors", t.Name(), len(errors))
	}

	format.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()

	treeStack, failedPaths, err := filetree.StackTreeRange(result.RefTrees, 0, 0)
	if len(failedPaths) > 0 {
		t.Errorf("expected no filepath errors, got %d", len(failedPaths))
	}
	if err != nil {
		t.Fatalf("%s: unable to stack trees: %v", t.Name(), err)
	}
	vm, err := NewFileTreeViewModel(treeStack, result.RefTrees, cache)
	if err != nil {
		t.Fatalf("%s: unable to create tree ViewModel: %+v", t.Name(), err)
	}
	return vm
}

func runTestCase(t *testing.T, vm *FileTree, width, height int, filterRegex *regexp.Regexp) {
	err := vm.Update(filterRegex, width, height)
	if err != nil {
		t.Errorf("failed to update viewmodel: %v", err)
	}

	err = vm.Render()
	if err != nil {
		t.Errorf("failed to render viewmodel: %v", err)
	}

	assertTestData(t, vm.Buffer.Bytes())
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

	// collapse /bin
	err := vm.ToggleCollapse(nil)
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
	err = vm.ToggleCollapse(nil)
	checkError(t, err, "unable to collapse /etc")

	runTestCase(t, vm, width, height, nil)
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
	err := vm.ToggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the next layer, compareMode = layer
	err = vm.SetTreeByLayer(0, 0, 1, 1)
	if err != nil {
		t.Errorf("unable to SetTreeByLayer: %v", err)
	}
	runTestCase(t, vm, width, height, nil)
}

func TestFileShowAggregateChanges(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse(nil)
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
	err := vm.ToggleCollapse(nil)
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
	err = vm.ToggleCollapse(nil)
	checkError(t, err, "unable to collapse /etc")

	// expand /etc
	err = vm.CursorRight(nil)
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

func TestFileTreeHideAddedRemovedModified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.ToggleCollapse(nil)
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
	err := vm.ToggleCollapse(nil)
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
	err := vm.ToggleCollapse(nil)
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
