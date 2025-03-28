package runtime

import (
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/runtime/ci"
	"github.com/wagoodman/dive/runtime/ui/key"
)

type Options struct {
	// analysis
	Image     string
	Source    dive.ImageSource
	BuildArgs []string

	// gating
	Ci           bool
	CiRules      []ci.Rule
	IgnoreErrors bool
	ExportFile   string

	// ui
	KeyBindings key.Bindings
}
