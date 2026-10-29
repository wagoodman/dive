package payload

import "github.com/wagoodman/dive/dive"

// Explore is the ExploreAnalysis payload.
type Explore struct {
	Analysis *dive.Analysis
	// Done is called once the UI no longer needs the analysis (its image can then be cleaned up)
	Done func()
}
