package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FilterView struct {
	*tview.InputField
}

func NewFilterView() *FilterView {
	inputField := tview.NewInputField()
	inputField.SetBackgroundColor(tcell.ColorGray)
	inputField.SetFieldTextColor(tcell.ColorBlack)
	inputField.SetFieldBackgroundColor(tcell.ColorGray)
	//inputField.SetPlaceholderTextColor(tcell.ColorBlack)
	inputField.SetLabelColor(tcell.ColorBlack)
	inputField.SetLabel("Path Filter: ")
	//inputField.SetPlaceholder("(regex)" )
	return &FilterView{
		InputField: inputField,
	}
}

func (fv *FilterView) Empty() bool {
	return fv.GetText() == ""
}