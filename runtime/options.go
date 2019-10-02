package runtime

import (
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive"
)

type Options struct {
	Ci         bool
	ImageId    string
	Engine     dive.Engine
	ExportFile string
	CiConfig   *viper.Viper
	BuildArgs  []string
}
