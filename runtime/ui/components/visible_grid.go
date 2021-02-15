package components

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
)

type GridPrimitive interface {
	VisiblePrimitive
	GetKeyBindings() []helpers.KeyBindingDisplay
}

type flexItem struct {
	Item       GridPrimitive // The item to be positioned. May be nil for an empty item.
	FixedSize  int           // The item's fixed size which may not be changed, 0 if it has no fixed size.
	Proportion int           // The item's proportion
	Focus      bool          // Whether or not this item attracts the layout's focus.
}

type VisibleFlex struct {
	*tview.Box

	// The items to be positioned.
	items   []*flexItem
	consume [][]int

	// FlexRow or FlexColumn.
	direction    int
	visible      VisibleFunc
	bindingArray []helpers.KeyBinding
}

func NewVisibleFlex() *VisibleFlex {
	return &VisibleFlex{
		Box:       tview.NewBox().SetBackgroundColor(tcell.ColorDefault),
		direction: tview.FlexColumn,
		visible:   Always(true),
	}
}

func (f *VisibleFlex) GetKeyBindings() []helpers.KeyBindingDisplay {
	result := []helpers.KeyBindingDisplay{}

	for _, binding := range f.bindingArray {
		result = append(result, helpers.KeyBindingDisplay{KeyBinding: &binding, Selected: AlwaysFalse, Hide: AlwaysTrue})
	}

	for _, item := range f.items {
		if item.Item.HasFocus() {
			result = append(result, item.Item.GetKeyBindings()...)
		}
	}

	return result
}

func (f *VisibleFlex) SetVisibility(visibleFunc VisibleFunc) GridPrimitive {
	f.visible = visibleFunc
	return f
}

func (f *VisibleFlex) SetDirection(direction int) *VisibleFlex {
	f.direction = direction
	return f
}

func (f *VisibleFlex) AddItem(item GridPrimitive, fixedSize, proportion int, focus bool) *VisibleFlex {
	f.items = append(f.items, &flexItem{Item: item, FixedSize: fixedSize, Proportion: proportion, Focus: focus})
	f.consume = append(f.consume, []int{})
	return f
}

// RemoveItem removes all items for the given primitive from the container,
// keeping the order of the remaining items intact.
func (f *VisibleFlex) RemoveItem(p GridPrimitive) *VisibleFlex {
	for index := len(f.items) - 1; index >= 0; index-- {
		if f.items[index].Item == p {
			f.items = append(f.items[:index], f.items[index+1:]...)
			f.consume = append(f.consume[:index], f.consume[index+1:]...)
		}
	}
	return f
}

func (f *VisibleFlex) Clear() *VisibleFlex {
	f.items = nil
	f.consume = [][]int{}
	return f
}

func (f *VisibleFlex) ResizeItem(p tview.Primitive, fixedSize, proportion int) *VisibleFlex {
	for _, item := range f.items {
		if item.Item == p {
			item.FixedSize = fixedSize
			item.Proportion = proportion
		}
	}
	return f
}

// TODO: update the  API here this is pretty rough
// Method provided to give configuration that would otherwise not be possible when primitives are repeated
func (f *VisibleFlex) SetConsumersByIndex(p GridPrimitive, consumeIndicies []int) *VisibleFlex {
	for i, item := range f.items {
		if item.Item == p {
			f.consume[i] = consumeIndicies
		}
	}
	return f
}

// TODO: update the  API here this is pretty rough
// Implementation notes:
// we want a list of indicies []int{} where each visible primitive corresponds to the first matching primitive
// in our list of items
func (f *VisibleFlex) SetConsumers(p GridPrimitive, consumes ...GridPrimitive) *VisibleFlex {
	indexMap := map[GridPrimitive]int{}
	for _, item := range f.items {
		_, ok := indexMap[item.Item]
		if !ok {
			indexMap[item.Item] = len(indexMap)
		}
	}

	consumeIndicies := []int{}
	for _, consumee := range consumes {
		if idx, ok := indexMap[consumee]; ok {
			consumeIndicies = append(consumeIndicies, idx)
		}
	}

	for i, item := range f.items {
		if item.Item == p {
			f.consume[i] = consumeIndicies
		}
	}
	return f
}

// Implementation notes:
// do not allow hidden items to recieve focus...  How would focus and vsisiblity be intertwined otherwise???
//   cases: i) A hidden element recieves focus (we can disallow this)
//          ii) a focused item becomes hidden (this is handled by individual element)
//
// This function prohibits case (i) above
func (f *VisibleFlex) Focus(delegate func(p tview.Primitive)) {
	for _, item := range f.items {
		if item.Item != nil && item.Focus && item.Item.Visible() {
			delegate(item.Item)
			return
		}
	}
}

////
//// Getters
////

// TODO: replace me with a focusable??
func (f *VisibleFlex) HasFocus() bool {
	for _, item := range f.items {
		if item.Item != nil && item.Item.HasFocus() {
			return true
		}
	}
	return false
}

func (f *VisibleFlex) Visible() bool {
	return f.visible(f)
}

////
//// Handlers
////

// Implementation notes:
// Should hidden elements be able to handle ( & consume ) mouse inputs??
// seems like the logical answer is no....
func (f *VisibleFlex) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return f.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !f.InRect(event.Position()) {
			return false, nil
		}

		// Pass mouse events along to the first child item that takes it.
		for _, item := range f.items {
			if item.Item == nil {
				continue
			}
			consumed, capture = item.Item.MouseHandler()(action, event, setFocus)
			if consumed {
				return
			}
		}

		return
	})
}

func (f *VisibleFlex) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		for _, item := range f.items {
			if item.Item != nil && item.Item.HasFocus() {
				if handler := item.Item.InputHandler(); handler != nil {
					handler(event, setFocus)
					return
				}
			}
		}
	})
}

func (f *VisibleFlex) Draw(screen tcell.Screen) {
	// skip drawing if grid is not visible
	x, y, totalWidth, totalHeight := f.GetInnerRect()
	hiddenFill(screen, f.GetBackgroundColor(), x, y, totalWidth, totalHeight)
	f.Box.Draw(screen)
	if !f.Visible() {
		return
	}
	// calculate a value to scale proportions by to avoid proportion rounding errors
	// (this happens when a item of proportion 2 is consumed by 3 other items)
	consumeLCM := lcm(lens(f.consume)...)

	// Calculate size and position of the items

	// How much space can we distribute?
	distSize := totalWidth
	if f.direction == tview.FlexRow {
		distSize = totalHeight
	}
	var proportionSum int
	for _, item := range f.items {
		if item.FixedSize > 0 {
			distSize -= item.FixedSize
		} else {
			proportionSum += item.Proportion * consumeLCM
		}
	}

	pos := x
	if f.direction == tview.FlexRow {
		pos = y
	}
	// go through assign sizes and check if visible
	proportionDelta := make([]int, len(f.items))
	fixedSizeDelta := make([]int, len(f.items))
	proportionLeft := proportionSum
	distLeft := distSize
	for i, item := range f.items {
		size := item.FixedSize
		if size <= 0 {
			if proportionLeft > 0 {
				size = distLeft * item.Proportion * consumeLCM / proportionLeft
				distLeft -= size
				proportionLeft -= (item.Proportion * consumeLCM)
			} else {
				size = 0
			}
		}

		if item.Item != nil {
			if f.direction == tview.FlexColumn {
				item.Item.SetRect(pos, y, size, totalHeight)
			} else {
				item.Item.SetRect(x, pos, totalWidth, size)
			}

			// now lets check if we are hidden as size may change this function call
			if !item.Item.Visible() && len(f.consume[i]) > 0 {
				denom := intMax(len(f.consume[i]), 1)
				proportionValue := item.Proportion * consumeLCM / denom
				for _, j := range f.consume[i] {
					proportionDelta[j] += proportionValue
				}

				div := item.FixedSize / denom
				mod := item.FixedSize % denom
				for _, j := range f.consume[i] {
					fixedSizeDelta[j] += div
					if j < mod {
						fixedSizeDelta[j] += 1
					}
				}
			}
		}
		pos += size
	}
	// go through assign sizes and check if visible
	proportionLeft = proportionSum
	distLeft = distSize
	// second pass where we actually update our views
	pos = x
	if f.direction == tview.FlexRow {
		pos = y
	}
	for i, item := range f.items {
		size := item.FixedSize + fixedSizeDelta[i]
		adjustedProportion := (item.Proportion * consumeLCM) + proportionDelta[i]
		if proportionLeft > 0 && item.Item.Visible() {
			// actually quite nice how this is going to end up perfectly filling the screen
			sizeFromProportion := (distLeft * adjustedProportion) / proportionLeft
			distLeft -= sizeFromProportion
			size += sizeFromProportion
			proportionLeft -= adjustedProportion
		}
		if item.Item != nil && item.Item.Visible() {
			if f.direction == tview.FlexColumn {
				item.Item.SetRect(pos, y, size, totalHeight)
			} else {
				item.Item.SetRect(x, pos, totalWidth, size)
			}
			// only update pos if we draw this item
			pos += size
		}
		if item.Item != nil && item.Item.Visible() {
			switch {
			case item.Item.HasFocus():
				defer item.Item.Draw(screen)
			case item.Item.Visible():
				item.Item.Draw(screen)
			}
		}
	}
}

// helpers

func hiddenFill(screen tcell.Screen, bgColor tcell.Color, x, y, width, height int) {
	// Fill background.
	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(bgColor)
	for curY := y; curY < y+height; curY++ {
		for curX := x; curX < x+width; curX++ {
			screen.SetContent(curX, curY, ' ', nil, background)
		}
	}
}

func lens(arr [][]int) []int {
	result := make([]int, len(arr))
	for i := 0; i < len(arr); i++ {
		result[i] = len(arr[i])
	}

	return result
}

func lcm(vals ...int) int {
	curLCM := 1
	maxVal := intMax(vals...)
	limit := int(math.Ceil(math.Sqrt(float64(maxVal)) + 1))
	div := 2
	for div <= limit {
		divFound := false
		for i, val := range vals {
			if val != 0 && val%div == 0 {
				divFound = true
				vals[i] = val / div
			}
		}
		if divFound {
			curLCM *= div
		} else {
			div++
		}
	}

	for _, val := range vals {
		if val != 0 {
			curLCM *= val
		}
	}
	return curLCM
}

func intMax(vals ...int) int {
	max := vals[0]
	for _, val := range vals {
		if max < val {
			max = val
		}
	}

	return max
}
