package handler

import (
	"errors"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

var _ progress.StagedProgressable = (*failableProgress)(nil)

// failableProgress wraps the progress of a stereoscope row so the row can be failed when its image source gives up.
// stereoscope marks a pull status complete on failure too, so the row would otherwise say it pulled the image.
type failableProgress struct {
	progress.Progressable
	err atomic.Pointer[error]
}

// failable wraps prog and tracks it as part of the current image source attempt
func (m *Handler) failable(prog progress.Progressable) *failableProgress {
	f := &failableProgress{Progressable: prog}
	m.attempt = append(m.attempt, f)
	return f
}

func (f *failableProgress) fail(err error) {
	f.err.CompareAndSwap(nil, &err)
}

func (f *failableProgress) Error() error {
	if err := f.err.Load(); err != nil {
		return *err
	}
	return f.Progressable.Error()
}

// a failed row is complete: zero of zero done
func (f *failableProgress) Size() int64 {
	if f.err.Load() != nil {
		return 0
	}
	return f.Progressable.Size()
}

func (f *failableProgress) Current() int64 {
	if f.err.Load() != nil {
		return 0
	}
	return f.Progressable.Current()
}

func (f *failableProgress) Stage() string {
	if s, ok := f.Progressable.(progress.Stager); ok {
		return s.Stage()
	}
	return ""
}

// handleSourceFailed fails every row of the attempt that just failed, the next source starts with a clean slate
func (m *Handler) handleSourceFailed(e partybus.Event) ([]tea.Model, tea.Cmd) {
	err := e.Error
	if err == nil {
		err = errors.New("image source failed")
	}
	for _, f := range m.attempt {
		f.fail(err)
	}
	m.attempt = nil
	return nil, nil
}
