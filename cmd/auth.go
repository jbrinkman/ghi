// Package cmd implements the commands for the GitHub Info CLI
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure authentication settings",
	Long:  `The 'auth' command configures authentication settings for database connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the subcommands: 'set' or 'info'")
		cmd.Help()
	},
}

var authSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set authentication settings",
	Long:  `Set authentication settings for database connection.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, _ := cmd.Flags().GetString("db-url")
		authToken, _ := cmd.Flags().GetString("auth-token")
		username, _ := cmd.Flags().GetString("username")

		if dbURL == "" && authToken == "" && username == "" {
			log.Fatal("At least one flag must be provided")
		}

		// Create .ghi directory in user's home directory if it doesn't exist
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home directory: %v", err)
		}

		configDir := filepath.Join(home, ".ghi")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0700); err != nil {
				log.Fatalf("Error creating config directory: %v", err)
			}
		}

		// Create or append to the .env file
		envFile := filepath.Join(configDir, ".env")
		f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("Error opening env file: %v", err)
		}
		defer f.Close()

		// Write settings to file and set environment variables
		if dbURL != "" {
			if _, err := f.WriteString(fmt.Sprintf("GHI_DB_URL=%s\n", dbURL)); err != nil {
				log.Fatalf("Error writing db URL: %v", err)
			}
			os.Setenv("GHI_DB_URL", dbURL)
			fmt.Println("Database URL set successfully")
		}

		if authToken != "" {
			if _, err := f.WriteString(fmt.Sprintf("GHI_AUTH_TOKEN=%s\n", authToken)); err != nil {
				log.Fatalf("Error writing auth token: %v", err)
			}
			os.Setenv("GHI_AUTH_TOKEN", authToken)
			fmt.Println("Auth token set successfully")
		}

		if username != "" {
			if _, err := f.WriteString(fmt.Sprintf("GHI_USERNAME=%s\n", username)); err != nil {
				log.Fatalf("Error writing username: %v", err)
			}
			os.Setenv("GHI_USERNAME", username)
			fmt.Println("Username set successfully")
		}
	},
}

var authInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current authentication settings",
	Long:  `Display current authentication settings used for database connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read environment variables
		dbURL := os.Getenv("GHI_DB_URL")
		authToken := os.Getenv("GHI_AUTH_TOKEN")
		username := os.Getenv("GHI_USERNAME")

		fmt.Println("Current Authentication Settings:")
		if dbURL != "" {
			fmt.Printf("Database URL: %s\n", dbURL)
		} else {
			fmt.Println("Database URL: Not set")
		}

		if authToken != "" {
			fmt.Println("Auth Token: [Set]")
		} else {
			fmt.Println("Auth Token: Not set")
		}

		if username != "" {
			fmt.Printf("Username: %s\n", username)
		} else {
			fmt.Println("Username: Not set")
		}

		// Check for config file
		home, err := os.UserHomeDir()
		if err == nil {
			configFile := filepath.Join(home, ".ghi", ".env")
			if _, err := os.Stat(configFile); err == nil {
				fmt.Printf("\nCredentials file: %s\n", configFile)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authSetCmd)
	authCmd.AddCommand(authInfoCmd)

	// Define flags for authSetCmd
	authSetCmd.Flags().String("db-url", "", "Turso/LibSQL database URL")
	authSetCmd.Flags().String("auth-token", "", "Authentication token for database")
	authSetCmd.Flags().String("username", "", "Your username for review tracking")
}
