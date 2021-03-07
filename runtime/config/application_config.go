package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type ApplicationConfig struct {
	ConfigFile      string
	FileTree        fileTreeViewConfig `mapstructure:"filetree"`
	Layer           layerViewConfig    `mapstructure:"layer"`
	Keybinding      KeybindingConfig   `mapstructure:"keybinding"`
	Diff            diffConfig         `mapstructure:"diff"`
	Log             loggingConfig      `mapstructure:"log"`
	ContainerEngine string             `mapstructure:"container-engine"`
	IgnoreErrors    bool               `mapstructure:"ignore-errors"`
}

func LoadApplicationConfig(v *viper.Viper, cfgFile string) (*ApplicationConfig, error) {
	setDefaultConfigValues(v)
	readConfig(cfgFile)

	instance := &ApplicationConfig{}
	if err := v.Unmarshal(instance); err != nil {
		return nil, fmt.Errorf("unable to unmarshal application config: %w", err)
	}

	instance.ConfigFile = v.ConfigFileUsed()

	return instance, instance.build()
}

func (a *ApplicationConfig) build() error {
	// currently no other configs need to be built
	return a.Diff.build()
}

func (a ApplicationConfig) String() string {
	content, err := yaml.Marshal(&a)
	if err != nil {
		return "[no config]"
	}
	return string(content)
}

func readConfig(cfgFile string) {
	viper.SetEnvPrefix("DIVE")
	// replace all - with _ when looking for matching environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// if config files are present, load them
	if cfgFile == "" {
		// default configs are ignored if not found
		cfgFile = getDefaultCfgFile()
	}

	if cfgFile == "" {
		return
	}

	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("unable to read config file: %w", err))
	}
}

// getDefaultCfgFile checks for config file in paths from xdg specs
// and in $HOME/.config/dive/ directory
// defaults to $HOME/.dive.yaml
func getDefaultCfgFile() string {
	home, err := homedir.Dir()
	if err != nil {
		err = fmt.Errorf("unable to get home dir: %w", err)
		panic(err)
	}

	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	xdgDirs := os.Getenv("XDG_CONFIG_DIRS")
	xdgPaths := append([]string{xdgHome}, strings.Split(xdgDirs, ":")...)
	allDirs := append(xdgPaths, path.Join(home, ".config"))

	for _, val := range allDirs {
		file := findInPath(val)
		if len(file) > 0 {
			return file
		}
	}

	for _, altPath := range []string{path.Join(home, ".dive.yaml"), path.Join(home, ".dive.yml")} {
		if _, err := os.Stat(altPath); os.IsNotExist(err) {
			continue
		} else if err != nil {
			panic(err)
		}
		return altPath
	}
	return ""
}

// findInPath returns first "*.yaml" or "*.yml" file in path's subdirectory "dive"
// if not found returns empty string
func findInPath(pathTo string) string {
	directory := path.Join(pathTo, "dive")
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}

	for _, file := range files {
		filename := file.Name()
		if path.Ext(filename) == ".yaml" || path.Ext(filename) == ".yml" {
			return path.Join(directory, filename)
		}
	}
	return ""
}

func setDefaultConfigValues(v *viper.Viper) {
	// logging
	v.SetDefault("log.enabled", true)
	v.SetDefault("log.level", "debug")
	v.SetDefault("log.path", "./dive.log")
	// keybindings: status view / global
	v.SetDefault("keybinding.quit", "Ctrl+C")
	v.SetDefault("keybinding.toggle-view", "Tab")
	v.SetDefault("keybinding.filter-files", "Ctrl+f")
	// keybindings: layer view
	v.SetDefault("keybinding.compare-all", "Ctrl+A")
	v.SetDefault("keybinding.compare-layer", "Ctrl+L")
	// keybindings: filetree view
	v.SetDefault("keybinding.toggle-collapse-dir", "Space")
	v.SetDefault("keybinding.toggle-collapse-all-dir", "Ctrl+Space")
	v.SetDefault("keybinding.toggle-filetree-attributes", "Ctrl+B")
	v.SetDefault("keybinding.toggle-added-files", "Ctrl+A")
	v.SetDefault("keybinding.toggle-removed-files", "Ctrl+R")
	v.SetDefault("keybinding.toggle-modified-files", "Ctrl+M")
	v.SetDefault("keybinding.toggle-unmodified-files", "Ctrl+U")
	v.SetDefault("keybinding.toggle-wrap-tree", "Ctrl+P")
	v.SetDefault("keybinding.page-up", "PgUp")
	v.SetDefault("keybinding.page-down", "PgDn")
	// layer view
	v.SetDefault("layer.show-aggregated-changes", false)
	// filetree view
	v.SetDefault("diff.hide", "")
	v.SetDefault("filetree.collapse-dir", false)
	v.SetDefault("filetree.pane-width", 0.5)
	v.SetDefault("filetree.show-attributes", true)
	// general behavior
	v.SetDefault("container-engine", "docker")
	v.SetDefault("ignore-errors", false)
}
