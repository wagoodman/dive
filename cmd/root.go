package cmd

import (
	"fmt"
	"github.com/wagoodman/dive/utils"
	"os"

	"github.com/k0kubun/go-ansi"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dive [IMAGE]",
	Short: "Docker Image Visualizer & Explorer",
	Long: `This tool provides a way to discover and explore the contents of a docker image. Additionally the tool estimates
the amount of wasted space and identifies the offending files from the image.`,
	Args: cobra.MaximumNArgs(1),
	Run:  analyze,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		utils.Exit(1)
	}
}

func init() {
	ansi.CursorHide()

	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLogging)

	// TODO: add config options
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dive.yaml)")

	rootCmd.PersistentFlags().BoolP("version", "v", false, "display version number")
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

		// Search config in home directory with name ".dive" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dive")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// initLogging sets up the loggin object with a formatter and location
func initLogging() {
	// TODO: clean this up and make more configurable
	var filename string = "dive.log"
	// create the log file if doesn't exist. And append to it if it already exists.
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	Formatter := new(log.TextFormatter)
	Formatter.DisableTimestamp = true
	log.SetFormatter(Formatter)
	log.SetLevel(log.DebugLevel)
	if err != nil {
		// cannot open log file. Logging to stderr
		fmt.Println(err)
	} else {
		log.SetOutput(f)
	}
	log.Debug("Starting Dive...")
}
