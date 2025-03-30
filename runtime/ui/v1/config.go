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
	Image        string
	Content      ContentReader
	Analysis     *image.AnalysisResult
	KeyBindings  key.Bindings
	IgnoreErrors bool

	stack     filetree.Comparer
	stackErrs error
	do        *sync.Once
}

func (c *Config) TreeComparer() (filetree.Comparer, error) {
	if c.do == nil {
		c.do = &sync.Once{}
	}
	c.do.Do(func() {
		treeStack := filetree.NewComparer(c.Analysis.RefTrees)
		errs := treeStack.BuildCache()
		if errs != nil {
			if !c.IgnoreErrors {
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
