package cmd

import (
	"fmt"
	"github.com/wagoodman/dive/filetree"
	"github.com/wagoodman/dive/utils"
	"io/ioutil"
	"os"

	"github.com/k0kubun/go-ansi"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var exportFile string

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
	ansi.CursorHide()

	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dive.yaml, ~/.config/dive.yaml, or $XDG_CONFIG_HOME/dive.yaml)")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "display version number")

	rootCmd.Flags().StringVarP(&exportFile, "json", "j", "", "Skip the interactive TUI and write the layer analysis statistics to a given file.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			utils.Exit(1)
		}

		// allow for:
		//    ~/.dive.yaml
		//    $XDG_CONFIG_HOME/dive.yaml
		//    ~/.config/dive.yaml
		hiddenHomeCfg := home + string(os.PathSeparator) + ".dive.yaml"
		if _, err := os.Stat(hiddenHomeCfg); os.IsNotExist(err) {
			viper.SetConfigName("dive")

			xdgHome := os.Getenv("XDG_CONFIG_HOME")
			if len(xdgHome) > 0 {
				viper.AddConfigPath(xdgHome)
			}

			viper.AddConfigPath(home + string(os.PathSeparator) + ".config")

		} else {
			viper.SetConfigFile(hiddenHomeCfg)
		}
	}

	viper.SetDefault("log.level", log.InfoLevel.String())
	viper.SetDefault("log.path", "./dive.log")
	viper.SetDefault("log.enabled", true)
	// keybindings: status view / global
	viper.SetDefault("keybinding.quit", "ctrl+c")
	viper.SetDefault("keybinding.toggle-view", "tab, ctrl+space")
	viper.SetDefault("keybinding.filter-files", "ctrl+f, ctrl+slash")
	// keybindings: layer view
	viper.SetDefault("keybinding.compare-all", "ctrl+a")
	viper.SetDefault("keybinding.compare-layer", "ctrl+l")
	// keybindings: filetree view
	viper.SetDefault("keybinding.toggle-collapse-dir", "space")
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

	if viper.GetBool("log.enabled") == false {
		log.SetOutput(ioutil.Discard)
	}

	logFileObj, err := os.OpenFile(viper.GetString("log.path"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
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

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		log.SetOutput(logFileObj)
	}

	log.Debug("Starting Dive...")
}
