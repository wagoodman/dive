package v1

import (
	"errors"
	"fmt"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"github.com/wagoodman/dive/runtime/ui/v1/key"
	"sync"
)

type Config struct {
	// required input
	Analysis    image.Analysis
	Content     ContentReader
	Preferences Preferences

	stack     filetree.Comparer
	stackErrs error
	do        *sync.Once
}

type Preferences struct {
	KeyBindings                key.Bindings
	IgnoreErrors               bool
	ShowFiletreeAttributes     bool
	ShowAggregatedLayerChanges bool
	CollapseFiletreeDirectory  bool
	FiletreePaneWidth          float64
	FiletreeDiffHide           []string
}

func DefaultPreferences() Preferences {
	return Preferences{
		KeyBindings:                key.DefaultBindings(),
		ShowFiletreeAttributes:     true,
		ShowAggregatedLayerChanges: true,
		CollapseFiletreeDirectory:  false, // don't start with collapsed directories
		FiletreePaneWidth:          0.5,
		FiletreeDiffHide:           []string{}, // empty slice means show all
	}
}

func (c *Config) TreeComparer() (filetree.Comparer, error) {
	if c.do == nil {
		c.do = &sync.Once{}
	}
	c.do.Do(func() {
		treeStack := filetree.NewComparer(c.Analysis.RefTrees)
		errs := treeStack.BuildCache()
		if errs != nil {
			if !c.Preferences.IgnoreErrors {
				errs = append(errs, fmt.Errorf("file tree has path errors (use '--ignore-errors' to attempt to continue)"))
				c.stackErrs = errors.Join(errs...)
				return
			}
		}
		c.stack = treeStack
	})

	return c.stack, c.stackErrs
}

type ContentReader interface {
	Extract(id string, layer string, path string) error
}
