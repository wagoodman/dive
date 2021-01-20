package fakes

import (
	"regexp"
	"sync"
)

type FilterModel struct {
	GetFilterCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Regexp *regexp.Regexp
		}
		Stub func() *regexp.Regexp
	}
	SetFilterCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			R *regexp.Regexp
		}
		Stub func(*regexp.Regexp)
	}
}

func (f *FilterModel) GetFilter() *regexp.Regexp {
	f.GetFilterCall.Lock()
	defer f.GetFilterCall.Unlock()
	f.GetFilterCall.CallCount++
	if f.GetFilterCall.Stub != nil {
		return f.GetFilterCall.Stub()
	}
	return f.GetFilterCall.Returns.Regexp
}
func (f *FilterModel) SetFilter(param1 *regexp.Regexp) {
	f.SetFilterCall.Lock()
	defer f.SetFilterCall.Unlock()
	f.SetFilterCall.CallCount++
	f.SetFilterCall.Receives.R = param1
	if f.SetFilterCall.Stub != nil {
		f.SetFilterCall.Stub(param1)
	}
}
