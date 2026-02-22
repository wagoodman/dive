package view

import "testing"

func TestImageDetailsShouldRenderOnFirstRender(t *testing.T) {
	v := &ImageDetails{}

	if !v.shouldRender(120, false) {
		t.Fatalf("expected first render to be required")
	}
}

func TestImageDetailsShouldRenderWhenLayoutOrSelectionChanges(t *testing.T) {
	v := &ImageDetails{
		rendered:             true,
		lastRenderWidth:      120,
		lastRenderIsSelected: false,
	}

	if v.shouldRender(120, false) {
		t.Fatalf("expected render to be skipped when state has not changed")
	}

	if !v.shouldRender(121, false) {
		t.Fatalf("expected render when width changes")
	}

	if !v.shouldRender(120, true) {
		t.Fatalf("expected render when selection state changes")
	}
}
