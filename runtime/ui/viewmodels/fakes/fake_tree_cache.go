package fakes

import (
	"sync"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
)

type TreeCache struct {
	GetTreeCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Key filetree.TreeIndexKey
		}
		Returns struct {
			TreeModel viewmodels.TreeModel
			Error     error
		}
		Stub func(filetree.TreeIndexKey) (viewmodels.TreeModel, error)
	}
}

func (f *TreeCache) GetTree(param1 filetree.TreeIndexKey) (viewmodels.TreeModel, error) {
	f.GetTreeCall.Lock()
	defer f.GetTreeCall.Unlock()
	f.GetTreeCall.CallCount++
	f.GetTreeCall.Receives.Key = param1
	if f.GetTreeCall.Stub != nil {
		return f.GetTreeCall.Stub(param1)
	}
	return f.GetTreeCall.Returns.TreeModel, f.GetTreeCall.Returns.Error
}
