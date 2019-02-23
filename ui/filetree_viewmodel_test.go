package ui

import (
	"bytes"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/image"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
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

func initializeTestViewModel(t *testing.T) *FileTreeViewModel {
	result, err := image.TestLoadDockerImageTar("../.data/test-docker-image.tar")
	if err != nil {
		t.Fatalf("%s: unable to fetch analysis: %v", t.Name(), err)
	}
	cache := filetree.NewFileTreeCache(result.RefTrees)
	cache.Build()

	Formatting.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()

	return NewFileTreeViewModel(filetree.StackTreeRange(result.RefTrees, 0, 0), result.RefTrees, cache)
}

func runTestCase(t *testing.T, vm *FileTreeViewModel, width, height int, filterRegex *regexp.Regexp) {
	err := vm.Update(filterRegex, width, height)
	if err != nil {
		t.Errorf("failed to update viewmodel: %v", err)
	}

	err = vm.Render()
	if err != nil {
		t.Errorf("failed to render viewmodel: %v", err)
	}

	assertTestData(t, vm.mainBuf.Bytes())
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
	err := vm.toggleCollapse(nil)
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
	err = vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /etc")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeDirCollapseAll(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	err := vm.toggleCollapseAll()
	checkError(t, err, "unable to collapse all dir")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeSelectLayer(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the next layer, compareMode = layer
	err = vm.setTreeByLayer(0, 0, 1, 1)
	if err != nil {
		t.Errorf("unable to setTreeByLayer: %v", err)
	}
	runTestCase(t, vm, width, height, nil)
}

func TestFileShowAggregateChanges(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the next layer, compareMode = layer
	err = vm.setTreeByLayer(0, 0, 1, 13)
	checkError(t, err, "unable to setTreeByLayer")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreePageDown(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 10
	vm.Setup(0, height)
	vm.ShowAttributes = true
	vm.Update(nil, width, height)

	err := vm.PageDown()
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
	vm.Update(nil, width, height)

	err := vm.PageDown()
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
	err := vm.toggleCollapse(nil)
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
	err = vm.toggleCollapse(nil)
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
	err := vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.setTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to setTreeByLayer: %v", err)
	}

	// hide added files
	err = vm.toggleShowDiffType(filetree.Added)
	checkError(t, err, "unable hide added files")

	// hide modified files
	err = vm.toggleShowDiffType(filetree.Changed)
	checkError(t, err, "unable hide added files")

	// hide removed files
	err = vm.toggleShowDiffType(filetree.Removed)
	checkError(t, err, "unable hide added files")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideUnmodified(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.setTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to setTreeByLayer: %v", err)
	}

	// hide unmodified files
	err = vm.toggleShowDiffType(filetree.Unchanged)
	checkError(t, err, "unable hide added files")

	runTestCase(t, vm, width, height, nil)
}

func TestFileTreeHideTypeWithFilter(t *testing.T) {
	vm := initializeTestViewModel(t)

	width, height := 100, 100
	vm.Setup(0, height)
	vm.ShowAttributes = true

	// collapse /bin
	err := vm.toggleCollapse(nil)
	checkError(t, err, "unable to collapse /bin")

	// select the 7th layer, compareMode = layer
	err = vm.setTreeByLayer(0, 0, 1, 7)
	if err != nil {
		t.Errorf("unable to setTreeByLayer: %v", err)
	}

	// hide added files
	err = vm.toggleShowDiffType(filetree.Added)
	checkError(t, err, "unable hide added files")

	regex, err := regexp.Compile("saved")
	if err != nil {
		t.Errorf("could not create filter regex: %+v", err)
	}

	runTestCase(t, vm, width, height, regex)
}
