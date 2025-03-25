/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jbrinkman/ghi/pkg/clients"
	"github.com/jbrinkman/ghi/pkg/db"
	"github.com/jbrinkman/ghi/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reviewCmd represents the review command
var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "List pull requests you have reviewed",
	Long: `The 'review' command shows a list of pull requests you have reviewed.
You can filter by repository using the --repo flag.`,
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
		viper.BindPFlag("all", cmd.Flags().Lookup("all"))

		repo := viper.GetString("repo")
		all := viper.GetBool("all")

		logger.Debug("Command arguments: %v", args)
		logger.Debug("Repository filter: %s", repo)
		logger.Debug("Show all flag: %v", all)

		// Check for username
		username := os.Getenv("GHI_USERNAME")
		if username == "" {
			log.Fatal("GHI_USERNAME environment variable not set. Use 'ghi auth set --username YOUR_USERNAME' to set it")
		}

		// Connect to database
		dbClient, err := db.NewClient()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer dbClient.Close()

		// Initialize schema if needed
		ctx := context.Background()
		if err := dbClient.InitSchema(ctx); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}

		// Get reviews
		reviews, err := dbClient.GetReviewsByReviewer(ctx, username, repo)
		if err != nil {
			log.Fatalf("Failed to fetch reviews: %v", err)
		}

		// Create a GitHub client to fetch PR status
		client, err := clients.NewGitHubClient()
		if err != nil {
			log.Fatalf("Failed to create GitHub client: %v", err)
		}

		// Print reviews in a table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Repository\tPR Number\tStatus\tReviewed At")
		fmt.Fprintln(w, "----------\t---------\t------\t-----------")

		for _, review := range reviews {
			// Parse repository to get owner and repo name
			parts := strings.Split(review.Repo, "/")
			if len(parts) != 2 {
				logger.Debug("Invalid repository format: %s", review.Repo)
				continue
			}
			owner, repoName := parts[0], parts[1]

			// Get PR status
			pr, _, err := client.PullRequests.Get(ctx, owner, repoName, review.PRNumber)
			if err != nil {
				logger.Debug("Failed to fetch PR status for %s #%d: %v",
					review.Repo, review.PRNumber, err)
				// If we can't fetch the PR and --all is not set, skip it
				if !all {
					continue
				}
				fmt.Fprintf(w, "%s\t#%d\t%s\t%s\n",
					review.Repo,
					review.PRNumber,
					"unknown",
					review.Timestamp.Format(time.RFC822))
				continue
			}

			// Skip closed PRs unless --all is set
			if !all && *pr.State == "closed" {
				continue
			}

			fmt.Fprintf(w, "%s\t#%d\t%s\t%s\n",
				review.Repo,
				review.PRNumber,
				*pr.State,
				review.Timestamp.Format(time.RFC822))
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	// Define flags
	reviewCmd.Flags().StringP("repo", "r", "", "Filter reviews by repository (owner/repo)")
	reviewCmd.Flags().BoolP("all", "a", false, "Show all reviews, including closed PRs")
	reviewCmd.Flags().StringP("config", "c", "", "Path to the configuration file")
}
