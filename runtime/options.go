package runtime

import (
	"github.com/spf13/viper"
)

type Options struct {
	Ci         bool
	ImageId    string
	ExportFile string
	CiConfig   *viper.Viper
	BuildArgs  []string
}
