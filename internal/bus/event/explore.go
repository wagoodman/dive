package event

import "github.com/wagoodman/dive/dive"

// Explore is the ExploreAnalysis payload.
// todo: move back to the payload package once dive/image/docker (which imports payload) is deleted, payload
// importing dive is an import cycle until then
type Explore struct {
	Analysis *dive.Analysis
	// Done is called once the UI no longer needs the analysis (its image can then be cleaned up)
	Done func()
}
