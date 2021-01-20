package viewmodels_test

import (
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
	"regexp"
	"testing"
)

func TestFilterModel(t *testing.T) {

	testNilFilterView(t)
	testReFilterView(t)
}


func testNilFilterView(t *testing.T) {
	filterView := viewmodels.NewFilterViewModel(nil)
	filter := filterView.GetFilter()
	if filter != nil {
		t.Errorf("expected nil got %#v", filter)
	}
}

func testReFilterView(t *testing.T) {
	filterView := viewmodels.NewFilterViewModel(nil)

	r := regexp.MustCompile("some regex")
	filterView.SetFilter(r)
	filter := filterView.GetFilter()
	if filter != r {
		t.Errorf("expected %q got %#v", r, filter)
	}
	if filter.String() != "some regex" {
		t.Errorf("expected 'some regex' got %s", filter.String())
	}
}