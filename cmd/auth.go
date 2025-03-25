/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication settings",
	Long: `The auth command allows you to manage authentication settings.
For GitHub access, you'll need to set your token to avoid rate limiting.
You can create a token at https://github.com/settings/tokens`,
}

var authSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set authentication settings",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		token, _ := cmd.Flags().GetString("token")
		dburl, _ := cmd.Flags().GetString("db-url")
		dbtoken, _ := cmd.Flags().GetString("db-token")

		// Create config directory if it doesn't exist
		configDir := filepath.Join(os.Getenv("HOME"), ".ghi")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Error creating config directory: %v", err)
		}

		// Update environment file
		envFile := filepath.Join(configDir, "env")
		env := make(map[string]string)

		// Read existing env file if it exists
		if data, err := os.ReadFile(envFile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				}
			}
		}

		// Update values
		if username != "" {
			env["GHI_USERNAME"] = username
		}
		if token != "" {
			env["GHI_GITHUB_TOKEN"] = token
		}
		if dburl != "" {
			env["GHI_DB_URL"] = dburl
		}
		if dbtoken != "" {
			env["GHI_AUTH_TOKEN"] = dbtoken
		}

		// Write back to file
		f, err := os.Create(envFile)
		if err != nil {
			log.Fatalf("Error creating env file: %v", err)
		}
		defer f.Close()

		for k, v := range env {
			fmt.Fprintf(f, "%s=%s\n", k, v)
		}

		fmt.Println("Authentication settings updated successfully")
		if token != "" {
			fmt.Println("GitHub token set - API requests will now use authenticated rate limits (5000/hour)")
		}
	},
}

var authShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current authentication settings",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Username: %s\n", os.Getenv("GHI_USERNAME"))

		// Don't show the full token for security
		token := os.Getenv("GHI_GITHUB_TOKEN")
		if token != "" {
			fmt.Printf("GitHub Token: %s...%s\n", token[:4], token[len(token)-4:])
		} else {
			fmt.Println("GitHub Token: not set")
		}

		fmt.Printf("Database URL: %s\n", os.Getenv("GHI_DB_URL"))

		dbToken := os.Getenv("GHI_AUTH_TOKEN")
		if dbToken != "" {
			fmt.Printf("Database Token: %s...%s\n", dbToken[:4], dbToken[len(dbToken)-4:])
		} else {
			fmt.Println("Database Token: not set")
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authSetCmd)
	authCmd.AddCommand(authShowCmd)

	// Add flags for auth set command
	authSetCmd.Flags().StringP("username", "u", "", "Your GitHub username")
	authSetCmd.Flags().StringP("token", "t", "", "Your GitHub personal access token")
	authSetCmd.Flags().String("db-url", "", "Database URL")
	authSetCmd.Flags().String("db-token", "", "Database authentication token")
}
