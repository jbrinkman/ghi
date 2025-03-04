/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Version information
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghi",
	Short: "GitHub Info - A command line tool for GitHub repository information",
	Long: `GitHub Info (ghi) provides a simple command line interface for 
retrieving and displaying information about GitHub repositories, 
including pull requests and repository statistics.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver, comm, dt string) {
	version = ver
	commit = comm
	date = dt

	// Load environment variables from ~/.ghi/.env if it exists
	loadEnvFile()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// loadEnvFile loads environment variables from ~/.ghi/.env file if it exists
func loadEnvFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		return // Silently fail if we can't get home dir
	}

	envFile := filepath.Join(home, ".ghi", ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return // File doesn't exist, that's okay
	}

	// Read and parse the .env file
	file, err := os.Open(envFile)
	if err != nil {
		return // Can't open file, just skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first equals sign only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = fmt.Sprintf("%s (Built: %s, Commit: %s)", version, date, commit)

	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.github-info.yaml)")

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
