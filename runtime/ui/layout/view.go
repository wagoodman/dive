package layout

import (
	"github.com/wagoodman/dive/runtime/ui/view"
)

type View interface {
	view.Identifiable
	view.Dimensional
}
