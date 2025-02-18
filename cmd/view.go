package cmd

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View details of a specific pull request",
	Long:  `The 'view' command retrieves and displays details of a specific pull request from a specified GitHub repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				log.Fatalf("Error reading config file: %v", err)
			}
		}

		// Bind flags to viper
		viper.BindPFlag("repo", cmd.Flags().Lookup("repo"))
		viper.BindPFlag("number", cmd.Flags().Lookup("number"))
		viper.BindPFlag("web", cmd.Flags().Lookup("web"))

		repo := viper.GetString("repo")
		if repo == "" {
			log.Fatal("The --repo flag is required")
		}

		number := viper.GetInt("number")
		if number == 0 {
			log.Fatal("The --number flag is required")
		}

		web := viper.GetBool("web")

		// Split the repo into owner and repo name
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			log.Fatal("Invalid repository format. Use 'owner/repo'")
		}
		owner, repoName := parts[0], parts[1]

		// Create a new Github client
		ctx := context.Background()
		client := github.NewClient(nil)

		// Get the pull request details
		pr, _, err := client.PullRequests.Get(ctx, owner, repoName, number)
		if err != nil {
			log.Fatalf("Error fetching pull request #%d: %v", number, err)
		}

		if web {
			openBrowser(*pr.HTMLURL)
			return
		}

		// Print the pull request details
		fmt.Printf("Pull Request #%d\n", *pr.Number)
		fmt.Printf("Title: %s\n", *pr.Title)
		fmt.Printf("Author: %s\n", *pr.User.Login)
		fmt.Printf("State: %s\n", *pr.State)
		fmt.Printf("Created At: %s\n", pr.CreatedAt.Format(time.RFC1123))
		fmt.Printf("Updated At: %s\n", pr.UpdatedAt.Format(time.RFC1123))
		if pr.MergedAt != nil {
			fmt.Printf("Merged At: %s\n", pr.MergedAt.Format(time.RFC1123))
		}
		fmt.Printf("URL: %s\n", *pr.HTMLURL)
		fmt.Printf("Body:\n%s\n", *pr.Body)
	},
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Fatalf("Failed to open browser: %v", err)
	}
}

func init() {
	prCmd.AddCommand(viewCmd)

	// Define the --repo flag for viewCmd
	viewCmd.Flags().StringP("repo", "r", "", "The name of the Github repository (owner/repo)")
	viewCmd.MarkFlagRequired("repo")

	// Define the --number flag for viewCmd
	viewCmd.Flags().IntP("number", "n", 0, "The number of the pull request")
	viewCmd.MarkFlagRequired("number")

	// Define the --config flag for viewCmd
	viewCmd.Flags().StringP("config", "c", "", "Path to the configuration file")

	// Define the --web flag for viewCmd
	viewCmd.Flags().BoolP("web", "w", false, "Open the pull request in the default web browser")
}
