package v1

import (
	"errors"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/format"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/ui/v1/key"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/dive/image"
	"golang.org/x/net/context"
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

type Colors struct {
	Selected   *color.Color
	Added      *color.Color
	Removed    *color.Color
	Modified   *color.Color
	Unmodified *color.Color
	Layer      *color.Color
	Header     *color.Color
}

func (c Colors) Apply() {
	if c.Selected != nil {
		format.SetSelected(c.Selected)
	}
	if c.Layer != nil {
		format.SetLayer(c.Layer)
	}
	if c.Header != nil {
		format.SetHeader(c.Header)
	}
	if c.Added != nil {
		filetree.SetDiffTypeColor(filetree.Added, c.Added)
	}
	if c.Removed != nil {
		filetree.SetDiffTypeColor(filetree.Removed, c.Removed)
	}
	if c.Modified != nil {
		filetree.SetDiffTypeColor(filetree.Modified, c.Modified)
	}
	if c.Unmodified != nil {
		filetree.SetDiffTypeColor(filetree.Unmodified, c.Unmodified)
	}
}

func ResetDefaultColors() {
	format.ResetDefaults()
	filetree.ResetDiffTypeColors()
}

type Preferences struct {
	KeyBindings                key.Bindings
	IgnoreErrors               bool
	ShowFiletreeAttributes     bool
	ShowAggregatedLayerChanges bool
	CollapseFiletreeDirectory  bool
	FiletreePaneWidth          float64
	FiletreeDiffHide           []string
	Colors                     Colors
}

func DefaultPreferences() Preferences {
	return Preferences{
		KeyBindings:                key.DefaultBindings(),
		ShowFiletreeAttributes:     true,
		ShowAggregatedLayerChanges: true,
		CollapseFiletreeDirectory:  false, // don't start with collapsed directories
		FiletreePaneWidth:          0.5,
		FiletreeDiffHide:           []string{}, // empty slice means show all
		Colors:                     Colors{},
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
	Extract(ctx context.Context, id string, layer string, path string) error
}
