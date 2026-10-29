package payload

import (
	"github.com/wagoodman/go-progress"
)

type GenericTask struct {
	// required fields

	Title Title

	// optional format fields

	HideOnSuccess      bool
	HideStageOnSuccess bool

	// optional fields

	ID       string
	ParentID string
	Context  string
}

type GenericProgress struct {
	*progress.AtomicStage
	*progress.Manual
}

type Title struct {
	Default      string
	WhileRunning string
	OnSuccess    string
}
