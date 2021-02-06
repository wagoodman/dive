package components

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/components/helpers"
)

type DiveApplication struct {
	*tview.Application

	boundList []BoundView

	bindings []helpers.KeyBinding
}

func NewDiveApplication(app *tview.Application) *DiveApplication {
	return &DiveApplication{
		Application: app,
		boundList:   []BoundView{},
	}
}

func (d *DiveApplication) GetKeyBindings() []helpers.KeyBindingDisplay {
	result := []helpers.KeyBindingDisplay{}
	for i := 0; i < len(d.bindings); i++ {
		binding := d.bindings[i]
		logrus.Debug(fmt.Sprintf("adding keybinding with name %s", binding.Display))
		result = append(result, helpers.KeyBindingDisplay{KeyBinding: &binding, Selected: AlwaysFalse, Hide: AlwaysFalse})
	}

	for _, bound := range d.boundList {
		if bound.HasFocus() {
			result = append(result, bound.GetKeyBindings()...)
		}
	}

	return result
}

func (d *DiveApplication) AddBindings(bindings ...helpers.KeyBinding) *DiveApplication {
	d.bindings = append(d.bindings, bindings...)

	return d
}

func (d *DiveApplication) AddBoundViews(views ...BoundView) *DiveApplication {
	d.boundList = append(d.boundList, views...)

	return d
}

// Application always has focus
func (d *DiveApplication) HasFocus() bool {
	return true
}
