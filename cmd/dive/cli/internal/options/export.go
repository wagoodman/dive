package options

import (
	"fmt"
	"os"
	"path"

	"github.com/anchore/clio"
)

var _ interface {
	clio.FlagAdder
	clio.PostLoader
} = (*Export)(nil)

// Export provides configuration for data export functionality
type Export struct {
	// Path to export analysis results as JSON (empty string = disabled)
	JsonPath string `yaml:"json-path" json:"json-path" mapstructure:"json-path"`
}

func DefaultExport() Export {
	return Export{}
}

func (o *Export) AddFlags(flags clio.FlagSet) {
	flags.StringVarP(&o.JsonPath, "json", "j", "Skip the interactive TUI and write the layer analysis statistics to a given file.")
}

func (o *Export) PostLoad() error {

	if o.JsonPath != "" {
		dir := path.Dir(o.JsonPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("directory for JSON export does not exist: %s", dir)
		}
	}

	return nil
}
