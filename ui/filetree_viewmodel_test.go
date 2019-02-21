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
	"testing"
)

const doCapture = false

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func testCaseDataFilePath(name string) string {
	return filepath.Join("testdata", name + ".txt")
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := testCaseDataFilePath(name)
	theBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to load test data ('%s'): %+v", name, err)
	}
	return theBytes
}

func helperCaptureBytes(t *testing.T, name string, data []byte) {
	if !doCapture {
		t.Fatalf("cannot capture data in test mode: %s", name)
	}

	path := testCaseDataFilePath(name)
	err := ioutil.WriteFile(path, data, 0644)

	if err != nil {
		t.Fatalf("unable to save test data ('%s'): %+v", name, err)
	}
}

func helperCheckDiff(t *testing.T, testCase string, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf(dmp.DiffPrettyText(diffs))
		t.Errorf("%s: bytes mismatch", testCase)
	}
}

func helperAssertTestData(t *testing.T, name string, actualBytes []byte) {
	path := testCaseDataFilePath(name)
	if !fileExists(path) {
		helperCaptureBytes(t, name, actualBytes)
	}
	expectedBytes := helperLoadBytes(t, name)
	helperCheckDiff(t, name, expectedBytes, actualBytes)
}

func TestFileTreeGoCase(t *testing.T) {
	testCase := "FileTreeGoCase"
	result, err := image.TestLoadDockerImageTar("../.data/test-docker-image.tar")
	if err != nil {
		t.Fatalf("Test_Export: unable to fetch analysis: %v", err)
	}
	cache := filetree.NewFileTreeCache(result.RefTrees)
	cache.Build()

	// :(
	Formatting.Selected = color.New(color.ReverseVideo, color.Bold).SprintFunc()

	vm := NewFileTreeViewModel(filetree.StackTreeRange(result.RefTrees, 0, 0), result.RefTrees, cache)
	vm.Setup(0, 1000)
	vm.ShowAttributes = true

	err = vm.Update(nil, 100, 1000)
	if err != nil {
		t.Errorf("failed to update viewmodel: %v", err)
	}

	err = vm.Render()
	if err != nil {
		t.Errorf("failed to render viewmodel: %v", err)
	}

	helperAssertTestData(t, testCase, vm.mainBuf.Bytes())

}
