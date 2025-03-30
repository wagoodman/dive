package view

import "github.com/wagoodman/dive/runtime/ui/v1/viewmodel"

type LayerChangeListener func(viewmodel.LayerSelection) error
