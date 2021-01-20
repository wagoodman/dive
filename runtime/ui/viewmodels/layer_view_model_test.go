package viewmodels_test

import (
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"testing"
)

func TestLayersViewModel(t *testing.T) {
	testMode(t)
	testIndicies(t)
	testGetCurrentLayer(t)

	testGetPrintableLayers(t)
}

func testMode(t *testing.T) {
	lvm := viewmodels.NewLayersViewModel([]*image.Layer{})

	curMode := lvm.GetMode()
	if curMode != viewmodels.CompareSingleLayer {
		t.Errorf("expected %v got %v", viewmodels.CompareSingleLayer, curMode)
	}

	if err := lvm.SwitchLayerMode(); err != nil {
		t.Errorf("expected 'nil' got %q", err)
	}

	curMode = lvm.GetMode()
	if curMode != viewmodels.CompareAllLayers {
		t.Errorf("expected %v got %v", viewmodels.CompareAllLayers, curMode)
	}
}

func testIndicies(t *testing.T) {
	testCompareSingleIndicies(t)
	testCompareAllIndicies(t)
}
func testCompareSingleIndicies(t *testing.T) {
	layers := []*image.Layer{
		{
			Id:      "some-id",
		},
		{
			Id:      "some-id2",
		},
		{
			Id:      "some-id3",
		},
	}


	lvm := viewmodels.NewLayersViewModel(layers)

	cmpIndex := lvm.GetCompareIndicies()
	if cmpIndex != filetree.NewTreeIndexKey(0,0,0,0) {
		t.Errorf("expected index key {0,0,0,0}, got %#v", cmpIndex)
	}

	lvm.SetLayerIndex(2)

	cmpIndex = lvm.GetCompareIndicies()
	if cmpIndex != filetree.NewTreeIndexKey(0,1,2,2) {
		t.Errorf("expected index key {0,1,2,2}, got %#v", cmpIndex)
	}
}

func testCompareAllIndicies(t *testing.T) {
	layers := []*image.Layer{
		{
			Id:      "some-id",
		},
		{
			Id:      "some-id2",
		},
		{
			Id:      "some-id3",
		},
	}


	lvm := viewmodels.NewLayersViewModel(layers)
	err := lvm.SwitchLayerMode()
	errorCheck(t, err)

	if lvm.GetMode() != viewmodels.CompareAllLayers {
		t.Errorf("expected CompareAllLayers mode %d, got %d", viewmodels.CompareAllLayers, lvm.GetMode())
	}

	cmpIndex := lvm.GetCompareIndicies()
	if cmpIndex != filetree.NewTreeIndexKey(0,0,1,0) {
		t.Errorf("expected index key {0,0,0,0}, got %#v", cmpIndex)
	}

	lvm.SetLayerIndex(2)

	cmpIndex = lvm.GetCompareIndicies()
	if cmpIndex != filetree.NewTreeIndexKey(0,0,1,2) {
		t.Errorf("expected index key {0,0,1,2}, got %#v", cmpIndex)
	}
}

func testGetCurrentLayer(t *testing.T) {
	firstLayer := &image.Layer{
			Id:      "some-id",
	}
	secondLayer := &image.Layer{
		Id:      "some-id2",
	}
	layers := []*image.Layer{firstLayer, secondLayer}
	lvm := viewmodels.NewLayersViewModel(layers)

	if lvm.GetCurrentLayer() != firstLayer {
		t.Errorf("expected %#v, got %#v", *firstLayer, *(lvm.GetCurrentLayer()))
	}

	lvm.SetLayerIndex(1)

	if lvm.GetCurrentLayer() != secondLayer {
		t.Errorf("expected %#v, got %#v", *secondLayer, *(lvm.GetCurrentLayer()))
	}
}

func testGetPrintableLayers(t *testing.T) {
	layers := []*image.Layer{
		{
			Id:      "some-id",
			Index: 0,
			Command: "layer1 cmd",
			Size:    100,
			Tree:    nil,
			Names:   []string{"name1", "name2"},
			Digest:  "digest:layer1",

		},
		{
			Id:      "some-id2",
			Index: 1,
			Command: "layer2 cmd",
			Size:    200,
			Tree:    nil,
			Names:   []string{"name3", "name4"},
			Digest:  "digest:layer2",
		},
	}

	lvm := viewmodels.NewLayersViewModel(layers)
	printableLayers := lvm.GetPrintableLayers()

	if len(printableLayers) != 2 {
		t.Errorf("expected 2 got %d", len(printableLayers))
	}

	expectedFirstLayer := "  100 B  FROM some-id"
	if printableLayers[0].String() != expectedFirstLayer {
		t.Errorf("expected %s got %s", expectedFirstLayer, printableLayers[0].String())
	}

	expectedSecondLayer := "  200 B  layer2 cmd"
	if printableLayers[1].String() != expectedSecondLayer {
		t.Errorf("expected %s got %s", expectedSecondLayer, printableLayers[1].String())
	}


}


