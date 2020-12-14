package components

import (
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FilterModel interface {
	SetFilter(*regexp.Regexp)
	GetFilter() *regexp.Regexp
}

type FilterView struct {
	*tview.InputField
	FilterModel
}

func NewFilterView(filterModel FilterModel) *FilterView {
	inputField := tview.NewInputField()
	return &FilterView{
		InputField:  inputField,
		FilterModel: filterModel,
	}
}

func (fv *FilterView) Setup() *FilterView {
	fv.SetBackgroundColor(tcell.ColorGray)
	fv.SetFieldTextColor(tcell.ColorBlack)
	fv.SetFieldBackgroundColor(tcell.ColorGray)
	fv.SetLabelColor(tcell.ColorBlack)
	fv.SetLabel("Path Filter: ")

	fv.SetChangedFunc(
		func(textToCheck string) {
			var filterRegex *regexp.Regexp = nil
			var err error

			if len(textToCheck) > 0 {
				filterRegex, err = regexp.Compile(textToCheck)
				if err != nil {
					return
				}
			}
			fv.FilterModel.SetFilter(filterRegex)
			return
		})
	return fv
}

func (ll *FilterView) getBox() *tview.Box {
	return ll.Box
}

func (ll *FilterView) getDraw() drawFn {
	return ll.Draw
}

func (ll *FilterView) getInputWrapper() inputFn {
	return ll.InputHandler
}

func (fv *FilterView) Empty() bool {
	return fv.GetText() == ""
}
