package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/utils"
)

var cfgFile string
var exportFile string
var ciConfigFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dive [IMAGE]",
	Short: "Docker Image Visualizer & Explorer",
	Long: `This tool provides a way to discover and explore the contents of a docker image. Additionally the tool estimates
the amount of wasted space and identifies the offending files from the image.`,
	Args: cobra.MaximumNArgs(1),
	Run:  doAnalyzeCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		utils.Exit(1)
	}
	utils.Cleanup()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dive.yaml, ~/.config/dive/*.yaml, or $XDG_CONFIG_HOME/dive.yaml)")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "display version number")

	rootCmd.Flags().StringVarP(&exportFile, "json", "j", "", "Skip the interactive TUI and write the layer analysis statistics to a given file.")
	rootCmd.Flags().StringVar(&ciConfigFile, "ci-config", ".dive-ci", "If CI=true in the environment, use the given yaml to drive validation rules.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	filepathToCfg := getCfgFile(cfgFile)
	viper.SetConfigFile(filepathToCfg)

	viper.SetDefault("log.level", log.InfoLevel.String())
	viper.SetDefault("log.path", "./dive.log")
	viper.SetDefault("log.enabled", true)
	// keybindings: status view / global
	viper.SetDefault("keybinding.quit", "ctrl+c")
	viper.SetDefault("keybinding.toggle-view", "tab")
	viper.SetDefault("keybinding.filter-files", "ctrl+f, ctrl+slash")
	// keybindings: layer view
	viper.SetDefault("keybinding.compare-all", "ctrl+a")
	viper.SetDefault("keybinding.compare-layer", "ctrl+l")
	// keybindings: filetree view
	viper.SetDefault("keybinding.toggle-collapse-dir", "space")
	viper.SetDefault("keybinding.toggle-collapse-all-dir", "ctrl+space")
	viper.SetDefault("keybinding.toggle-filetree-attributes", "ctrl+b")
	viper.SetDefault("keybinding.toggle-added-files", "ctrl+a")
	viper.SetDefault("keybinding.toggle-removed-files", "ctrl+r")
	viper.SetDefault("keybinding.toggle-modified-files", "ctrl+m")
	viper.SetDefault("keybinding.toggle-unchanged-files", "ctrl+u")
	viper.SetDefault("keybinding.page-up", "pgup")
	viper.SetDefault("keybinding.page-down", "pgdn")

	viper.SetDefault("diff.hide", "")

	viper.SetDefault("layer.show-aggregated-changes", false)

	viper.SetDefault("filetree.collapse-dir", false)
	viper.SetDefault("filetree.pane-width", 0.5)
	viper.SetDefault("filetree.show-attributes", true)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// set global defaults (for performance)
	filetree.GlobalFileTreeCollapse = viper.GetBool("filetree.collapse-dir")
}

// initLogging sets up the logging object with a formatter and location
func initLogging() {
	var logFileObj *os.File
	var err error

	if viper.GetBool("log.enabled") {
		logFileObj, err = os.OpenFile(viper.GetString("log.path"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	Formatter := new(log.TextFormatter)
	Formatter.DisableTimestamp = true
	log.SetFormatter(Formatter)

	level, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	log.SetLevel(level)
	log.SetOutput(logFileObj)
	log.Debug("Starting Dive...")
}

// getCfgFile checks for config file in paths from xdg specs
// and in $HOME/.config/dive/ directory
// defaults to $HOME/.dive.yaml
func getCfgFile(fromFlag string) string {
	if fromFlag != "" {
		return fromFlag
	}

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		utils.Exit(0)
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
	return path.Join(home, ".dive.yaml")
}

// findInPath returns first "*.yaml" file in path's subdirectory "dive"
// if not found returns empty string
func findInPath(pathTo string) string {
	directory := path.Join(pathTo, "dive")
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}

	for _, file := range files {
		filename := file.Name()
		if path.Ext(filename) == ".yaml" {
			return path.Join(directory, filename)
		}
	}
	return ""
}
