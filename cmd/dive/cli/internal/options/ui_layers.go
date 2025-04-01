package options

import "github.com/anchore/clio"

var _ clio.FieldDescriber = (*UILayers)(nil)

// UILayers provides configuration for layer display behavior
type UILayers struct {
	ShowAggregatedChanges bool `yaml:"show-aggregated-changes" mapstructure:"show-aggregated-changes"`
}

func DefaultUILayers() UILayers {
	return UILayers{
		ShowAggregatedChanges: false,
	}
}

func (c *UILayers) DescribeFields(descriptions clio.FieldDescriptionSet) {
	descriptions.Add(&c.ShowAggregatedChanges, "show aggregated changes across all previous layers")
}
