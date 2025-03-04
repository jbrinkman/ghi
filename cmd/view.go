package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/jbrinkman/ghi/pkg/db"
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
		viper.BindPFlag("log", cmd.Flags().Lookup("log"))

		repo := viper.GetString("repo")
		if repo == "" {
			log.Fatal("The --repo flag is required")
		}

		number := viper.GetInt("number")
		if number == 0 {
			log.Fatal("The --number flag is required")
		}

		web := viper.GetBool("web")
		logReview := viper.GetBool("log")

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

		// Log the review if requested
		if logReview {
			if err := logPRReview(ctx, repo, number); err != nil {
				log.Printf("Warning: Failed to log review: %v", err)
			} else {
				fmt.Println("✅ Review logged successfully")
			}
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

		// If requested to log review, also show previous reviews
		if logReview {
			showPreviousReviews(ctx, repo, number)
		}
	},
}

// logPRReview logs a code review to the database
func logPRReview(ctx context.Context, repo string, prNumber int) error {
	// Check for username
	username := os.Getenv("GHI_USERNAME")
	if username == "" {
		return fmt.Errorf("GHI_USERNAME environment variable not set. Use 'ghi auth set --username YOUR_USERNAME' to set it")
	}

	// Connect to database
	dbClient, err := db.NewClient()
	if err != nil {
		return err
	}
	defer dbClient.Close()

	// Initialize schema if needed
	if err := dbClient.InitSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Log the review
	if err := dbClient.LogReview(ctx, repo, prNumber, username); err != nil {
		return err
	}

	return nil
}

// showPreviousReviews displays previous reviews for this PR
func showPreviousReviews(ctx context.Context, repo string, prNumber int) {
	dbClient, err := db.NewClient()
	if err != nil {
		fmt.Printf("\nCould not access review history: %v\n", err)
		return
	}
	defer dbClient.Close()

	reviews, err := dbClient.GetReviews(ctx, repo, prNumber)
	if err != nil {
		fmt.Printf("\nCould not fetch review history: %v\n", err)
		return
	}

	if len(reviews) == 0 {
		fmt.Println("\nNo previous reviews found for this PR")
		return
	}

	fmt.Println("\nReview History:")
	fmt.Println("----------------")
	for _, review := range reviews {
		fmt.Printf("- %s by %s at %s\n",
			repo,
			review.Reviewer,
			review.Timestamp.Format(time.RFC1123))
	}
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

	// Define the --log flag for viewCmd
	viewCmd.Flags().BoolP("log", "l", false, "Log that you are reviewing this PR")
}
