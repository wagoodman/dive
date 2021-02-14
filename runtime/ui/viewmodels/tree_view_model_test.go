package viewmodels_test

import (
	tar "archive/tar"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"github.com/wagoodman/dive/runtime/ui/viewmodels/fakes"
)

func TestTreeViewModel(t *testing.T) {
	testStringBetween(t)
	testVisitDepthParentFirst(t)
	testVisitDepthChildFirst(t)
	testRemovePath(t)
	testVisibleSize(t)
	testSetFilter(t)
	testToggleHiddenFileType(t)
	testGetHiddenFileType(t)
	testSetLayerIndex(t)
	testSwitchLayerMode(t)
}

func testStringBetween(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel
	expectedString := "the string between"
	tModel.StringBetweenCall.Returns.String = expectedString

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	out := tvm.StringBetween(1, 2, true)
	if out != expectedString {
		t.Fatalf("expected: %s got: %s", expectedString, out)
	}

	if tModel.StringBetweenCall.CallCount != 1 {
		t.Error("expected StringBetween to be called on TreeModel")
	}

	if tModel.StringBetweenCall.Receives.Start != 1 {
		t.Fatalf("expected start to be passed through as 1, got %d", tModel.StringBetweenCall.Receives.Start)
	}

	if tModel.StringBetweenCall.Receives.Stop != 2 {
		t.Fatalf("expected start to be passed through as 2, got %d", tModel.StringBetweenCall.Receives.Stop)
	}

	if !tModel.StringBetweenCall.Receives.ShowAttributes {
		t.Fatalf("expected start to be passed through as true, got %t", tModel.StringBetweenCall.Receives.ShowAttributes)
	}
}

func testVisitDepthChildFirst(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	errorCheck(t, err)

	visitor := func(*filetree.FileNode) error { return nil }
	evaluator := func(*filetree.FileNode) bool { return true }
	err = tvm.VisitDepthChildFirst(visitor, evaluator)
	errorCheck(t, err)

	if tModel.VisitDepthChildFirstCall.CallCount != 1 {
		t.Fatalf("unexpected number of calls on TreeModel, expected 1 got %d", tModel.VisitDepthParentFirstCall.CallCount)
	}
}

func testVisitDepthParentFirst(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	errorCheck(t, err)

	visitor := func(*filetree.FileNode) error { return nil }
	evaluator := func(*filetree.FileNode) bool { return true }
	err = tvm.VisitDepthParentFirst(visitor, evaluator)
	errorCheck(t, err)

	if tModel.VisitDepthParentFirstCall.CallCount != 1 {
		t.Fatalf("unexpected number of calls on TreeModel, expected 1 got %d", tModel.VisitDepthParentFirstCall.CallCount)
	}
}

func testRemovePath(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	removePathArg := "/some/path"
	err = tvm.RemovePath(removePathArg)
	errorCheck(t, err)

	if removePathArg != tModel.RemovePathCall.Receives.Path {
		t.Fatalf("expected: %s recieved: %s", removePathArg, tModel.RemovePathCall.Receives.Path)
	}
}

func testVisibleSize(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel
	tModel.VisibleSizeCall.Returns.Int = 15

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	size := tvm.VisibleSize()

	if size != 15 {
		t.Fatalf("expected %d got %d", 15, size)
	}

	if tModel.VisibleSizeCall.CallCount != 1 {
		t.Fatalf("expected VisibleSize to be called 1 times, got %d", tModel.VisibleSizeCall.CallCount)
	}
}

func testSetFilter(t *testing.T) {
	fModel := viewmodels.NewFilterViewModel(nil)
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := filetree.NewFileTree()
	_, _, err := tModel.AddPath("/dirA/dirB/file", filetree.FileInfo{
		Path:     "/dirA/dirB/file",
		TypeFlag: tar.TypeReg,
		Size:     100,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	_, _, err = tModel.AddPath("/dirA/dirC/other-thing", filetree.FileInfo{
		Path:     "/dirA/dirC/other-thing",
		TypeFlag: tar.TypeReg,
		Size:     1000,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	_, _, err = tModel.AddPath("/dirA/dirB/other-file", filetree.FileInfo{
		Path:     "/dirA/dirB/other-file",
		TypeFlag: tar.TypeReg,
		Size:     1000,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	hiddenNodes, err := getHiddenNodes(tModel)
	errorCheck(t, err)
	assertEqual(t, 0, len(hiddenNodes))

	r := regexp.MustCompile("other-file")
	tvm.SetFilter(r)

	hiddenNodes, err = getHiddenNodes(tModel)
	errorCheck(t, err)
	assertEqual(t, 3, len(hiddenNodes))
	assertEqual(t, "file", hiddenNodes[0].Name)
	assertEqual(t, "dirC", hiddenNodes[1].Name)
	assertEqual(t, "other-thing", hiddenNodes[2].Name)

	// Check if directories where all children are hidden are hidden as well
	r = regexp.MustCompile("file")
	tvm.SetFilter(r)

	hiddenNodes, err = getHiddenNodes(tModel)
	errorCheck(t, err)
	assertEqual(t, 2, len(hiddenNodes))

	assertEqual(t, "dirC", hiddenNodes[0].Name)
	assertEqual(t, "other-thing", hiddenNodes[1].Name)
}

func testToggleHiddenFileType(t *testing.T) {
	fModel := viewmodels.NewFilterViewModel(nil)
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := filetree.NewFileTree()
	_, _, err := tModel.AddPath("/dirA/file", filetree.FileInfo{
		Path:     "/dirA/file",
		TypeFlag: tar.TypeReg,
		Size:     100,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	tModel.Root.Children["dirA"].Children["file"].Data.DiffType = filetree.Added
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	hiddenNodes, err := getHiddenNodes(tModel)
	errorCheck(t, err)
	assertEqual(t, 0, len(hiddenNodes))

	tvm.ToggleHiddenFileType(filetree.Added)

	hiddenNodes, err = getHiddenNodes(tModel)
	errorCheck(t, err)
	assertEqual(t, 2, len(hiddenNodes))
	assertEqual(t, "dirA", hiddenNodes[0].Name)
	assertEqual(t, "file", hiddenNodes[1].Name)
}

func testGetHiddenFileType(t *testing.T) {
	fModel := viewmodels.NewFilterViewModel(nil)
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if tvm.GetHiddenFileType(filetree.Added) {
		t.Fatalf("expected Added filetype to not be hidden by default")
	}

	if tvm.GetHiddenFileType(filetree.Modified) {
		t.Fatalf("expected Added filetype to not be hidden by default")
	}

	if tvm.GetHiddenFileType(filetree.Removed) {
		t.Fatalf("expected Added filetype to not be hidden by default")
	}

	if tvm.GetHiddenFileType(filetree.Unmodified) {
		t.Fatalf("expected Added filetype to not be hidden by default")
	}

	tvm.ToggleHiddenFileType(filetree.Added)
	if !tvm.GetHiddenFileType(filetree.Added) {
		t.Fatalf("expected Added filetype to be hidden after toggling")
	}

	tvm.ToggleHiddenFileType(filetree.Modified)
	if !tvm.GetHiddenFileType(filetree.Modified) {
		t.Fatalf("expected Added filetype to be hidden after toggling")
	}

	tvm.ToggleHiddenFileType(filetree.Removed)
	if !tvm.GetHiddenFileType(filetree.Removed) {
		t.Fatalf("expected Added filetype to be hidden after toggling")
	}

	tvm.ToggleHiddenFileType(filetree.Unmodified)
	if !tvm.GetHiddenFileType(filetree.Unmodified) {
		t.Fatalf("expected Added filetype to be hidden after toggling")
	}
}

func testSetLayerIndex(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	tModel := &fakes.TreeModel{}
	tCache.GetTreeCall.Returns.TreeModel = tModel

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	testIndex := 10
	tvm.SetLayerIndex(testIndex)
	assertEqual(t, testIndex, lModel.SetLayerIndexCall.Receives.Index)
}

func testSwitchLayerMode(t *testing.T) {
	fModel := &fakes.FilterModel{}
	lModel := &fakes.LayersModel{}
	tCache := &fakes.TreeCache{}
	firstTreeModel := filetree.NewFileTree()
	_, _, err := firstTreeModel.AddPath("/collapsed-dir/collapsed-file", filetree.FileInfo{
		Path:     "/collapsed-dir/collapsed-file",
		TypeFlag: tar.TypeReg,
		Size:     100,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	// Second tree has no collapsed or hidden values set
	secondTreeModel := firstTreeModel.Copy()

	firstTreeModel.Root.Children["collapsed-dir"].Data.ViewInfo.Collapsed = true

	_, _, err = secondTreeModel.AddPath("/visible/visible-file", filetree.FileInfo{
		Path:     "/visible/visible-file",
		TypeFlag: tar.TypeReg,
		Size:     100,
		Mode:     os.ModePerm,
		Uid:      200,
		Gid:      200,
		IsDir:    false,
	})
	errorCheck(t, err)

	collapsedNodes, err := getCollapsedNodes(secondTreeModel)
	errorCheck(t, err)
	if len(collapsedNodes) != 0 {
		t.Fatalf("expected no nodes to be collapsed got %v", collapsedNodes)
	}

	key := filetree.NewTreeIndexKey(1, 2, 3, 4)
	tCache.GetTreeCall.Stub = func(k filetree.TreeIndexKey) (viewmodels.TreeModel, error) {
		if k == key {
			return secondTreeModel, nil
		}

		return firstTreeModel, nil
	}

	tvm, err := viewmodels.NewTreeViewModel(tCache, lModel, fModel)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	lModel.GetCompareIndiciesCall.Returns.TreeIndexKey = filetree.NewTreeIndexKey(1, 2, 3, 4)
	err = tvm.SwitchLayerMode()
	errorCheck(t, err)

	collapsedNodes, err = getCollapsedNodes(secondTreeModel)
	errorCheck(t, err)
	assertEqual(t, 1, len(collapsedNodes))

	assertEqual(t, "collapsed-dir", collapsedNodes[0].Name)
}

func containsString(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}

func getHiddenNodes(tModel *filetree.FileTree) ([]*filetree.FileNode, error) {
	result := []*filetree.FileNode{}

	err := tModel.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.ViewInfo.Hidden {
			result = append(result, node)
		}
		return nil
	}, nil)

	return result, err
}

func getCollapsedNodes(tModel *filetree.FileTree) ([]*filetree.FileNode, error) {
	result := []*filetree.FileNode{}

	err := tModel.VisitDepthParentFirst(func(node *filetree.FileNode) error {
		if node.Data.ViewInfo.Collapsed {
			result = append(result, node)
		}
		return nil
	}, nil)

	return result, err
}

func errorCheck(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

//
// Assertion helpers
//

func assertEqual(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected %#v, got %#v", expected, actual)
	}
}

func assertSliceContains(t *testing.T, expected interface{}, container []interface{}) {
	t.Helper()
	for i := range container {
		if reflect.DeepEqual(container[i], expected) {
			return
		}
	}
	t.Fatalf("Expected value %v not found", expected)
}

