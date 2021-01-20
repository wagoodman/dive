package fakes

import (
	"sync"

	"github.com/wagoodman/dive/dive/filetree"
)

type TreeModel struct {
	RemovePathCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Error error
		}
		Stub func(string) error
	}
	StringBetweenCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Start          int
			Stop           int
			ShowAttributes bool
		}
		Returns struct {
			String string
		}
		Stub func(int, int, bool) string
	}
	VisibleSizeCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Int int
		}
		Stub func() int
	}
	VisitDepthChildFirstCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Visitor   filetree.Visitor
			Evaluator filetree.VisitEvaluator
		}
		Returns struct {
			Error error
		}
		Stub func(filetree.Visitor, filetree.VisitEvaluator) error
	}
	VisitDepthParentFirstCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Visitor   filetree.Visitor
			Evaluator filetree.VisitEvaluator
		}
		Returns struct {
			Error error
		}
		Stub func(filetree.Visitor, filetree.VisitEvaluator) error
	}
}

func (f *TreeModel) RemovePath(param1 string) error {
	f.RemovePathCall.Lock()
	defer f.RemovePathCall.Unlock()
	f.RemovePathCall.CallCount++
	f.RemovePathCall.Receives.Path = param1
	if f.RemovePathCall.Stub != nil {
		return f.RemovePathCall.Stub(param1)
	}
	return f.RemovePathCall.Returns.Error
}
func (f *TreeModel) StringBetween(param1 int, param2 int, param3 bool) string {
	f.StringBetweenCall.Lock()
	defer f.StringBetweenCall.Unlock()
	f.StringBetweenCall.CallCount++
	f.StringBetweenCall.Receives.Start = param1
	f.StringBetweenCall.Receives.Stop = param2
	f.StringBetweenCall.Receives.ShowAttributes = param3
	if f.StringBetweenCall.Stub != nil {
		return f.StringBetweenCall.Stub(param1, param2, param3)
	}
	return f.StringBetweenCall.Returns.String
}
func (f *TreeModel) VisibleSize() int {
	f.VisibleSizeCall.Lock()
	defer f.VisibleSizeCall.Unlock()
	f.VisibleSizeCall.CallCount++
	if f.VisibleSizeCall.Stub != nil {
		return f.VisibleSizeCall.Stub()
	}
	return f.VisibleSizeCall.Returns.Int
}
func (f *TreeModel) VisitDepthChildFirst(param1 filetree.Visitor, param2 filetree.VisitEvaluator) error {
	f.VisitDepthChildFirstCall.Lock()
	defer f.VisitDepthChildFirstCall.Unlock()
	f.VisitDepthChildFirstCall.CallCount++
	f.VisitDepthChildFirstCall.Receives.Visitor = param1
	f.VisitDepthChildFirstCall.Receives.Evaluator = param2
	if f.VisitDepthChildFirstCall.Stub != nil {
		return f.VisitDepthChildFirstCall.Stub(param1, param2)
	}
	return f.VisitDepthChildFirstCall.Returns.Error
}
func (f *TreeModel) VisitDepthParentFirst(param1 filetree.Visitor, param2 filetree.VisitEvaluator) error {
	f.VisitDepthParentFirstCall.Lock()
	defer f.VisitDepthParentFirstCall.Unlock()
	f.VisitDepthParentFirstCall.CallCount++
	f.VisitDepthParentFirstCall.Receives.Visitor = param1
	f.VisitDepthParentFirstCall.Receives.Evaluator = param2
	if f.VisitDepthParentFirstCall.Stub != nil {
		return f.VisitDepthParentFirstCall.Stub(param1, param2)
	}
	return f.VisitDepthParentFirstCall.Returns.Error
}
