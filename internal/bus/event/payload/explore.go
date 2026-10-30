package payload

import (
	"context"

	"github.com/wagoodman/dive/dive"
)

// Explore is the ExploreAnalysis payload.
type Explore struct {
	// Context is the command's, the UI quits when it is done
	Context  context.Context
	Analysis *dive.Analysis
	// Done is called once the UI no longer needs the analysis (its image can then be cleaned up)
	Done func()
}
