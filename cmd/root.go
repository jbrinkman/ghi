/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrinkman/ghi/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	// Version information
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghi",
	Short: "GitHub Information CLI provides tools for interacting with GitHub",
	Long: `GitHub Information CLI (ghi) provides tools for interacting with GitHub,
including viewing pull requests, managing reviews, and more.

Before using commands that interact with GitHub APIs, make sure to:
1. Set your GitHub username: ghi auth set --username YOUR_USERNAME
2. Set your GitHub token: ghi auth set --token YOUR_TOKEN

You can create a GitHub token at https://github.com/settings/tokens
Using a token increases your API rate limit from 60 to 5000 requests per hour.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Enable debug logging if flag is set
		debug := viper.GetBool("debug")
		if debug {
			logger.SetDebug(true)
			logger.Debug("Debug logging enabled")
		}

		// Load environment variables from .ghi/env file
		if envFile := filepath.Join(os.Getenv("HOME"), ".ghi", "env"); fileExists(envFile) {
			loadEnvFile(envFile)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver, comm, dt string) {
	version = ver
	commit = comm
	date = dt

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = fmt.Sprintf("%s (Built: %s, Commit: %s)", version, date, commit)

	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.github-info.yaml)")

	// Define debug flag with both long and short forms in a single call
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging to file")

	// Bind debug flag to viper
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".github-info" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".github-info")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// loadEnvFile loads environment variables from the specified file
func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		// Skip empty lines and comments
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		os.Setenv(key, value)
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
