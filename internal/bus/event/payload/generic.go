package payload

import (
	"context"
	"github.com/wagoodman/go-progress"
)

type genericProgressKey struct{}

func SetGenericProgressToContext(ctx context.Context, mon *GenericProgress) context.Context {
	return context.WithValue(ctx, genericProgressKey{}, mon)
}

func GetGenericProgressFromContext(ctx context.Context) *GenericProgress {
	mon, ok := ctx.Value(genericProgressKey{}).(*GenericProgress)
	if !ok {
		return nil
	}
	return mon
}

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
