// copied and adapted from syft v1.55.0 (github.com/anchore/syft, cmd/syft/cli/ui)

package handler

import "github.com/anchore/bubbly/bubbles/taskprogress"

func (m Handler) newTaskProgress(title taskprogress.Title, opts ...taskprogress.Option) taskprogress.Model {
	tsk := taskprogress.New(m.Running, opts...)

	tsk.HideProgressOnSuccess = true
	tsk.HideStageOnSuccess = true
	tsk.WindowSize = m.WindowSize
	tsk.TitleWidth = m.Config.TitleWidth
	tsk.TitleOptions = title

	if m.Config.AdjustDefaultTask != nil {
		tsk = m.Config.AdjustDefaultTask(tsk)
	}

	return tsk
}
