package layout

import (
	"testing"

	"github.com/awesome-gocui/gocui"
)

type testElement struct {
	t          *testing.T
	size       int
	layoutArea Area
	location   Location
}

func newTestElement(t *testing.T, size int, layoutArea Area, location Location) *testElement {
	return &testElement{
		t:          t,
		size:       size,
		layoutArea: layoutArea,
		location:   location,
	}
}

func (te *testElement) Name() string {
	return "dont care"
}
func (te *testElement) Layout(g *gocui.Gui, minX, minY, maxX, maxY int) error {
	actualLayoutArea := Area{
		minX: minX,
		minY: minY,
		maxX: maxX,
		maxY: maxY,
	}

	if te.layoutArea != actualLayoutArea {
		te.t.Errorf("expected layout area '%+v', got '%+v'", te.layoutArea, actualLayoutArea)
	}
	return nil
}
func (te *testElement) RequestedSize(available int) *int {
	if te.size == -1 {
		return nil
	}
	return &te.size
}
func (te *testElement) IsVisible() bool {
	return true
}
func (te *testElement) OnLayoutChange() error {
	return nil
}

type layoutReturn struct {
	area Area
	err  error
}

func Test_planAndLayoutHeaders(t *testing.T) {

	table := map[string]struct {
		headers  []*testElement
		expected layoutReturn
	}{
		"single header": {
			headers: []*testElement{newTestElement(t, 1, Area{
				minX: -1,
				minY: -1,
				maxX: 120,
				maxY: 0,
			}, LocationHeader)},
			expected: layoutReturn{
				area: Area{
					minX: -1,
					minY: 0,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
		"two headers": {
			headers: []*testElement{
				newTestElement(t, 1, Area{
					minX: -1,
					minY: -1,
					maxX: 120,
					maxY: 0,
				}, LocationHeader),
				newTestElement(t, 1, Area{
					minX: -1,
					minY: 0,
					maxX: 120,
					maxY: 1,
				}, LocationHeader),
			},
			expected: layoutReturn{
				area: Area{
					minX: -1,
					minY: 1,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
		"two odd-sized headers": {
			headers: []*testElement{
				newTestElement(t, 2, Area{
					minX: -1,
					minY: -1,
					maxX: 120,
					maxY: 1,
				}, LocationHeader),
				newTestElement(t, 3, Area{
					minX: -1,
					minY: 1,
					maxX: 120,
					maxY: 4,
				}, LocationHeader),
			},
			expected: layoutReturn{
				area: Area{
					minX: -1,
					minY: 4,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
	}

	for name, test := range table {
		t.Log("case: ", name, " ---")
		lm := NewManager()
		for _, element := range test.headers {
			lm.Add(element, element.location)
		}

		area, err := lm.planAndLayoutHeaders(nil, Area{
			minX: -1,
			minY: -1,
			maxX: 120,
			maxY: 80,
		})

		if err != test.expected.err {
			t.Errorf("%s: expected err '%+v', got error '%+v'", name, test.expected.err, err)
		}

		if area != test.expected.area {
			t.Errorf("%s: expected returned area '%+v', got area '%+v'", name, test.expected.area, area)
		}

	}
}

func Test_planAndLayoutColumns(t *testing.T) {

	table := map[string]struct {
		columns  []*testElement
		expected layoutReturn
	}{
		"single column": {
			columns: []*testElement{newTestElement(t, -1, Area{
				minX: -1,
				minY: -1,
				maxX: 120,
				maxY: 80,
			}, LocationColumn)},
			expected: layoutReturn{
				area: Area{
					minX: 120,
					minY: -1,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
		"two equal columns": {
			columns: []*testElement{
				newTestElement(t, -1, Area{
					minX: -1,
					minY: -1,
					maxX: 59,
					maxY: 80,
				}, LocationColumn),
				newTestElement(t, -1, Area{
					minX: 59,
					minY: -1,
					maxX: 119,
					maxY: 80,
				}, LocationColumn),
			},
			expected: layoutReturn{
				area: Area{
					minX: 119,
					minY: -1,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
		"two odd-sized columns": {
			columns: []*testElement{
				newTestElement(t, 30, Area{
					minX: -1,
					minY: -1,
					maxX: 29,
					maxY: 80,
				}, LocationColumn),
				newTestElement(t, -1, Area{
					minX: 29,
					minY: -1,
					maxX: 120,
					maxY: 80,
				}, LocationColumn),
			},
			expected: layoutReturn{
				area: Area{
					minX: 120,
					minY: -1,
					maxX: 120,
					maxY: 80,
				},
				err: nil,
			},
		},
	}

	for name, test := range table {
		t.Log("case: ", name, " ---")
		lm := NewManager()
		for _, element := range test.columns {
			lm.Add(element, element.location)
		}

		area, err := lm.planAndLayoutColumns(nil, Area{
			minX: -1,
			minY: -1,
			maxX: 120,
			maxY: 80,
		})

		if err != test.expected.err {
			t.Errorf("%s: expected err '%+v', got error '%+v'", name, test.expected.err, err)
		}

		if area != test.expected.area {
			t.Errorf("%s: expected returned area '%+v', got area '%+v'", name, test.expected.area, area)
		}

	}
}

func Test_layout(t *testing.T) {

	table := map[string]struct {
		elements []*testElement
	}{
		"1 header + 1 footer + 1 column": {
			elements: []*testElement{
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: -1,
						maxX: 120,
						maxY: 0,
					}, LocationHeader),
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: 78,
						maxX: 120,
						maxY: 80,
					}, LocationFooter),
				newTestElement(t, -1,
					Area{
						minX: -1,
						minY: 0,
						maxX: 120,
						maxY: 79,
					}, LocationColumn),
			},
		},
		"1 header + 1 footer + 3 column": {
			elements: []*testElement{
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: -1,
						maxX: 120,
						maxY: 0,
					}, LocationHeader),
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: 78,
						maxX: 120,
						maxY: 80,
					}, LocationFooter),
				newTestElement(t, -1,
					Area{
						minX: -1,
						minY: 0,
						maxX: 39,
						maxY: 79,
					}, LocationColumn),
				newTestElement(t, -1,
					Area{
						minX: 39,
						minY: 0,
						maxX: 79,
						maxY: 79,
					}, LocationColumn),
				newTestElement(t, -1,
					Area{
						minX: 79,
						minY: 0,
						maxX: 119,
						maxY: 79,
					}, LocationColumn),
			},
		},
		"1 header + 1 footer + 2 equal columns + 1 sized column": {
			elements: []*testElement{
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: -1,
						maxX: 120,
						maxY: 0,
					}, LocationHeader),
				newTestElement(t, 1,
					Area{
						minX: -1,
						minY: 78,
						maxX: 120,
						maxY: 80,
					}, LocationFooter),
				newTestElement(t, -1,
					Area{
						minX: -1,
						minY: 0,
						maxX: 19,
						maxY: 79,
					}, LocationColumn),
				newTestElement(t, 80,
					Area{
						minX: 19,
						minY: 0,
						maxX: 99,
						maxY: 79,
					}, LocationColumn),
				newTestElement(t, -1,
					Area{
						minX: 99,
						minY: 0,
						maxX: 119,
						maxY: 79,
					}, LocationColumn),
			},
		},
	}

	for name, test := range table {
		t.Log("case: ", name, " ---")
		lm := NewManager()
		for _, element := range test.elements {
			lm.Add(element, element.location)
		}

		err := lm.layout(nil, 120, 80)

		if err != nil {
			t.Fatalf("%s: unexpected error: %+v", name, err)
		}
	}
}
