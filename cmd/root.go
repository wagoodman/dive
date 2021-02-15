package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/filetree"
	"github.com/wagoodman/dive/internal/log"
	"github.com/wagoodman/dive/internal/logger"
	"github.com/wagoodman/dive/runtime"
	"github.com/wagoodman/dive/runtime/config"
)

var appConfig *config.ApplicationConfig
var cfgFile string
var exportFile string
var ciConfigFile string
var ciConfig = viper.New()
var isCi bool

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
		os.Exit(1)
	}
}

func init() {
	initCli()
	cobra.OnInitialize(initConfig)
}

func initCli() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dive.yaml, ~/.config/dive/*.yaml, or $XDG_CONFIG_HOME/dive.yaml)")
	rootCmd.PersistentFlags().String("source", "docker", "The container engine to fetch the image from. Allowed values: "+strings.Join(dive.ImageSources, ", "))
	rootCmd.PersistentFlags().BoolP("version", "v", false, "display version number")
	rootCmd.PersistentFlags().BoolP("ignore-errors", "i", false, "ignore image parsing errors and run the analysis anyway")
	rootCmd.Flags().BoolVar(&isCi, "ci", false, "Skip the interactive TUI and validate against CI rules (same as env var CI=true)")
	rootCmd.Flags().StringVarP(&exportFile, "json", "j", "", "Skip the interactive TUI and write the layer analysis statistics to a given file.")
	rootCmd.Flags().StringVar(&ciConfigFile, "ci-config", ".dive-ci", "If CI=true in the environment, use the given yaml to drive validation rules.")

	rootCmd.Flags().String("lowestEfficiency", "0.9", "(only valid with --ci given) lowest allowable image efficiency (as a ratio between 0-1), otherwise CI validation will fail.")
	rootCmd.Flags().String("highestWastedBytes", "disabled", "(only valid with --ci given) highest allowable bytes wasted, otherwise CI validation will fail.")
	rootCmd.Flags().String("highestUserWastedPercent", "0.1", "(only valid with --ci given) highest allowable percentage of bytes wasted (as a ratio between 0-1), otherwise CI validation will fail.")

	for _, key := range []string{"lowestEfficiency", "highestWastedBytes", "highestUserWastedPercent"} {
		if err := ciConfig.BindPFlag(fmt.Sprintf("rules.%s", key), rootCmd.Flags().Lookup(key)); err != nil {
			panic(fmt.Errorf("unable to bind '%s' flag: %v", key, err))
		}
	}

	if err := ciConfig.BindPFlag("ignore-errors", rootCmd.PersistentFlags().Lookup("ignore-errors")); err != nil {
		panic(fmt.Errorf("unable to bind 'ignore-errors' flag: %w", err))
	}

	if err := viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source")); err != nil {
		panic(fmt.Errorf("unable to bind 'source' flag: %w", err))
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error

	appConfig, err = config.LoadApplicationConfig(viper.GetViper(), cfgFile)
	if err != nil {
		panic(err)
	}

	// set global defaults
	filetree.GlobalFileTreeCollapse = appConfig.FileTree.CollapseDir
}

// initLogging sets up the logging object with a formatter and location
func initLogging() {
	logCfg := logger.LogrusConfig{
		EnableConsole: false,
		EnableFile:    appConfig.Log.Enabled,
		FileLocation:  appConfig.Log.Path,
		Level:         appConfig.Log.Level,
	}
	runtime.SetLogger(logger.NewLogrusLogger(logCfg))

	log.Debug("starting dive")
	log.Debug("config contents:\n", appConfig.String())
}
