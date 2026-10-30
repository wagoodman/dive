package view

import (
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/tui/viewmodel"
)

type LayerChangeListener func(viewmodel.LayerSelection) error
