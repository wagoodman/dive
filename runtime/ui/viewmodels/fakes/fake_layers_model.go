package fakes

import (
	"fmt"
	"sync"

	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/viewmodels"
)

type LayersModel struct {
	GetCompareIndiciesCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			TreeIndexKey filetree.TreeIndexKey
		}
		Stub func() filetree.TreeIndexKey
	}
	GetCurrentLayerCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Layer *image.Layer
		}
		Stub func() *image.Layer
	}
	GetModeCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			LayerCompareMode viewmodels.LayerCompareMode
		}
		Stub func() viewmodels.LayerCompareMode
	}
	GetPrintableLayersCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			StringerSlice []fmt.Stringer
		}
		Stub func() []fmt.Stringer
	}
	SetLayerIndexCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Index int
		}
		Returns struct {
			Bool bool
		}
		Stub func(int) bool
	}
	SwitchLayerModeCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Error error
		}
		Stub func() error
	}
}

func (f *LayersModel) GetCompareIndicies() filetree.TreeIndexKey {
	f.GetCompareIndiciesCall.Lock()
	defer f.GetCompareIndiciesCall.Unlock()
	f.GetCompareIndiciesCall.CallCount++
	if f.GetCompareIndiciesCall.Stub != nil {
		return f.GetCompareIndiciesCall.Stub()
	}
	return f.GetCompareIndiciesCall.Returns.TreeIndexKey
}
func (f *LayersModel) GetCurrentLayer() *image.Layer {
	f.GetCurrentLayerCall.Lock()
	defer f.GetCurrentLayerCall.Unlock()
	f.GetCurrentLayerCall.CallCount++
	if f.GetCurrentLayerCall.Stub != nil {
		return f.GetCurrentLayerCall.Stub()
	}
	return f.GetCurrentLayerCall.Returns.Layer
}
func (f *LayersModel) GetMode() viewmodels.LayerCompareMode {
	f.GetModeCall.Lock()
	defer f.GetModeCall.Unlock()
	f.GetModeCall.CallCount++
	if f.GetModeCall.Stub != nil {
		return f.GetModeCall.Stub()
	}
	return f.GetModeCall.Returns.LayerCompareMode
}
func (f *LayersModel) GetPrintableLayers() []fmt.Stringer {
	f.GetPrintableLayersCall.Lock()
	defer f.GetPrintableLayersCall.Unlock()
	f.GetPrintableLayersCall.CallCount++
	if f.GetPrintableLayersCall.Stub != nil {
		return f.GetPrintableLayersCall.Stub()
	}
	return f.GetPrintableLayersCall.Returns.StringerSlice
}
func (f *LayersModel) SetLayerIndex(param1 int) bool {
	f.SetLayerIndexCall.Lock()
	defer f.SetLayerIndexCall.Unlock()
	f.SetLayerIndexCall.CallCount++
	f.SetLayerIndexCall.Receives.Index = param1
	if f.SetLayerIndexCall.Stub != nil {
		return f.SetLayerIndexCall.Stub(param1)
	}
	return f.SetLayerIndexCall.Returns.Bool
}
func (f *LayersModel) SwitchLayerMode() error {
	f.SwitchLayerModeCall.Lock()
	defer f.SwitchLayerModeCall.Unlock()
	f.SwitchLayerModeCall.CallCount++
	if f.SwitchLayerModeCall.Stub != nil {
		return f.SwitchLayerModeCall.Stub()
	}
	return f.SwitchLayerModeCall.Returns.Error
}
