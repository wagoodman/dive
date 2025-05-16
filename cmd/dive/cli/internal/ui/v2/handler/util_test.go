package handler

import (
	"reflect"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func runModel(t testing.TB, m tea.Model, iterations int, message tea.Msg, h ...*sync.WaitGroup) tea.Model {
	t.Helper()
	if iterations == 0 {
		iterations = 1
	}
	m.Init()
	var cmd tea.Cmd = func() tea.Msg {
		return message
	}

	for _, each := range h {
		if each != nil {
			each.Wait()
		}
	}

	for i := 0; cmd != nil && i < iterations; i++ {
		msgs := flatten(cmd())
		var nextCmds []tea.Cmd
		var next tea.Cmd
		for _, msg := range msgs {
			t.Logf("Message: %+v %+v\n", reflect.TypeOf(msg), msg)
			m, next = m.Update(msg)
			nextCmds = append(nextCmds, next)
		}
		cmd = tea.Batch(nextCmds...)
	}

	return m
}

func flatten(ps ...tea.Msg) (msgs []tea.Msg) {
	for _, p := range ps {
		if bm, ok := p.(tea.BatchMsg); ok {
			for _, m := range bm {
				if m == nil {
					continue
				}
				msgs = append(msgs, flatten(m())...)
			}
		} else {
			msgs = []tea.Msg{p}
		}
	}
	return msgs
}
