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

	"github.com/google/go-github/v69/github"
	"github.com/jbrinkman/ghi/pkg/clients"
	"github.com/jbrinkman/ghi/pkg/db"
	"github.com/jbrinkman/ghi/pkg/logger"
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

		logger.Debug("Command arguments: %v", args)
		logger.Debug("Repository: %s, PR Number: %d", repo, number)
		logger.Debug("Web flag: %v, Log review flag: %v", web, logReview)

		// Split the repo into owner and repo name
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			log.Fatal("Invalid repository format. Use 'owner/repo'")
		}
		owner, repoName := parts[0], parts[1]

		// Create a GitHub client to fetch PR and issue data
		client, err := clients.NewGitHubClient()
		if err != nil {
			log.Fatalf("Failed to create GitHub client: %v", err)
		 }

		// Create context
		ctx := context.Background()

		// Fetch the PR data
		var pr *github.PullRequest

		// Get the pull request details
		logger.Debug("Fetching pull request details for %s/%s #%d", owner, repoName, number)
		pr, _, err = client.PullRequests.Get(ctx, owner, repoName, number)
		if err != nil {
			log.Fatalf("Error fetching pull request #%d: %v", number, err)
		}

		logger.Debug("Successfully retrieved PR #%d: %s", number, *pr.Title)

		// Log the review if requested
		if logReview {
			logger.Debug("Logging PR review for %s #%d", repo, number)
			if err := logPRReview(ctx, repo, number); err != nil {
				log.Printf("Warning: Failed to log review: %v", err)
			} else {
				fmt.Println("✅ Review logged successfully")
				logger.Debug("Review logged successfully")
			}
		}

		if web {
			logger.Debug("Opening PR in web browser: %s", *pr.HTMLURL)
			openBrowser(*pr.HTMLURL)
			return
		}

		// Print the pull request details
		fmt.Printf("Pull Request #%d\n", *pr.Number)

		// Add DRAFT: prefix to title if PR is in draft state
		title := *pr.Title
		if pr.GetDraft() {
			title = "DRAFT: " + title
		}

		fmt.Printf("Title: %s\n", title)
		fmt.Printf("Author: %s\n", *pr.User.Login)
		fmt.Printf("State: %s\n", *pr.State)

		// Add draft status - using GetDraft() directly with v69
		draftStatus := "[ ]"
		if pr.GetDraft() {
			draftStatus = "[X]" // Changed from "[✓]" to "[X]" to match reviewer indicator
			logger.Debug("PR #%d is a draft", *pr.Number)
		}
		fmt.Printf("Draft: %s\n", draftStatus)

		// Handle timestamps safely by checking if GetTime() returns nil
		if createdAt := pr.CreatedAt.GetTime(); createdAt != nil {
			fmt.Printf("Created At: %s\n", createdAt.Format(time.RFC1123))
		}

		if updatedAt := pr.UpdatedAt.GetTime(); updatedAt != nil {
			fmt.Printf("Updated At: %s\n", updatedAt.Format(time.RFC1123))
		}

		if pr.MergedAt != nil {
			if mergedAt := pr.MergedAt.GetTime(); mergedAt != nil {
				fmt.Printf("Merged At: %s\n", mergedAt.Format(time.RFC1123))
			}
		}

		fmt.Printf("URL: %s\n", *pr.HTMLURL)
		fmt.Printf("Body:\n%s\n", *pr.Body)

		// If requested to log review, also show previous reviews
		if logReview {
			logger.Debug("Showing previous reviews for PR #%d", number)
			showPreviousReviews(ctx, repo, number)
		}
	},
}

// logPRReview logs a code review to the database
func logPRReview(ctx context.Context, repo string, prNumber int) error {
	// Check for username
	username := os.Getenv("GHI_USERNAME")
	if username == "" {
		logger.Debug("GHI_USERNAME environment variable not set")
		return fmt.Errorf("GHI_USERNAME environment variable not set. Use 'ghi auth set --username YOUR_USERNAME' to set it")
	}

	logger.Debug("Logging review with username: %s", username)

	// Connect to database
	dbClient, err := db.NewClient()
	if err != nil {
		logger.Debug("Failed to connect to database: %v", err)
		return err
	}
	defer dbClient.Close()

	// Initialize schema if needed
	logger.Debug("Initializing database schema")
	if err := dbClient.InitSchema(ctx); err != nil {
		logger.Debug("Failed to initialize database schema: %v", err)
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Log the review
	logger.Debug("Writing review record to database")
	if err := dbClient.LogReview(ctx, repo, prNumber, username); err != nil {
		logger.Debug("Failed to log review: %v", err)
		return err
	}

	logger.Debug("Review logged successfully")
	return nil
}

// showPreviousReviews displays previous reviews for this PR
func showPreviousReviews(ctx context.Context, repo string, prNumber int) {
	dbClient, err := db.NewClient()
	if err != nil {
		fmt.Printf("\nCould not access review history: %v\n", err)
		logger.Debug("Failed to connect to database: %v", err)
		return
	}
	defer dbClient.Close()

	logger.Debug("Fetching reviews for %s #%d", repo, prNumber)
	reviews, err := dbClient.GetReviews(ctx, repo, prNumber)
	if err != nil {
		fmt.Printf("\nCould not fetch review history: %v\n", err)
		logger.Debug("Failed to fetch reviews: %v", err)
		return
	}

	if len(reviews) == 0 {
		fmt.Println("\nNo previous reviews found for this PR")
		logger.Debug("No reviews found for PR")
		return
	}

	logger.Debug("Found %d previous reviews", len(reviews))
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

	logger.Debug("Attempting to open URL: %s", url)
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
		logger.Debug("Failed to open browser: %v", err)
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
