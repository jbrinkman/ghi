// Package cmd implements the commands for the GitHub Info CLI
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/jbrinkman/ghi/pkg/db"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reviewCmd represents the review command
var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "List your PR reviews from a specified date range",
	Long: `The 'review' command retrieves and lists pull requests you've reviewed during a specified date range.
This data is pulled from the local Turso/LibSQL database that tracks your reviews when using the --log flag with the view command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		repo := viper.GetString("repo")
		startDateStr := viper.GetString("start-date")
		endDateStr := viper.GetString("end-date")

		// Parse dates
		startDate := time.Time{}
		endDate := time.Now()
		var err error

		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				log.Fatalf("Invalid start date format. Use YYYY-MM-DD: %v", err)
			}
		} else {
			// Default to 30 days ago if no start date provided
			startDate = time.Now().AddDate(0, 0, -30)
		}

		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				log.Fatalf("Invalid end date format. Use YYYY-MM-DD: %v", err)
			}
		}

		// Connect to database
		ctx := context.Background()
		dbClient, err := db.NewClient()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer dbClient.Close()

		// Get reviews from database
		reviews, err := dbClient.GetReviewsByDateRange(ctx, repo, startDate, endDate)
		if err != nil {
			log.Fatalf("Failed to get reviews: %v", err)
		}

		if len(reviews) == 0 {
			fmt.Printf("No reviews found between %s and %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			return
		}

		// Setup GitHub client to fetch PR details
		client := github.NewClient(nil)

		// Setup table for output
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		// Remove Repository column if filtering by specific repo
		if repo != "" {
			t.AppendHeader(table.Row{"Number", "Title", "Author", "State", "Review Date"})
		} else {
			t.AppendHeader(table.Row{"Number", "Title", "Author", "State", "Review Date", "Repository"})
		}

		// Create maps to track which PRs we've already processed
		// This helps avoid duplicate PRs in the table when you've reviewed the same PR multiple times
		processedPRs := make(map[string]bool)

		// Fetch and print PR details
		for _, review := range reviews {
			// Create a key to identify this PR
			prKey := fmt.Sprintf("%s-%d", review.Repo, review.PRNumber)

			// Skip if we've already processed this PR
			if processedPRs[prKey] {
				continue
			}
			processedPRs[prKey] = true

			// Split repo into owner/name
			parts := splitRepo(review.Repo)
			if len(parts) != 2 {
				// Skip invalid repos
				continue
			}
			owner, repoName := parts[0], parts[1]

			// Fetch PR details from GitHub
			pr, _, err := client.PullRequests.Get(ctx, owner, repoName, review.PRNumber)
			if err != nil {
				// If we can't get PR details, still show what we know
				if repo != "" {
					t.AppendRow(table.Row{
						strconv.Itoa(review.PRNumber),
						"[Unable to fetch PR]",
						"Unknown",
						"Unknown",
						review.Timestamp.Format("2006-01-02"),
					})
				} else {
					t.AppendRow(table.Row{
						strconv.Itoa(review.PRNumber),
						"[Unable to fetch PR]",
						"Unknown",
						"Unknown",
						review.Timestamp.Format("2006-01-02"),
						review.Repo,
					})
				}
				continue
			}

			// Format PR number
			prNumber := strconv.Itoa(review.PRNumber)

			// Trim the title if it's longer than 40 characters
			title := *pr.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			// Add row to table, with or without repo column based on filter
			if repo != "" {
				t.AppendRow(table.Row{
					prNumber,
					title,
					*pr.User.Login,
					*pr.State,
					review.Timestamp.Format("2006-01-02"),
				})
			} else {
				t.AppendRow(table.Row{
					prNumber,
					title,
					*pr.User.Login,
					*pr.State,
					review.Timestamp.Format("2006-01-02"),
					review.Repo,
				})
			}
		}

		// Print header with summary
		fmt.Println("=====================================")

		// Include repo name in header if filtering by repo
		if repo != "" {
			fmt.Printf("PR Reviews for %s\nbetween %s and %s\n",
				repo,
				startDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"))
		} else {
			fmt.Printf("PR Reviews between %s and %s\n",
				startDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"))
		}

		fmt.Printf("Count: %d\n", len(processedPRs))
		fmt.Println("=====================================")
		fmt.Println()

		// Render the table
		t.Render()
	},
}

// splitRepo breaks a repo string into owner and name
func splitRepo(repo string) []string {
	return strings.Split(repo, "/")
}

func init() {
	prCmd.AddCommand(reviewCmd)

	// Define flags
	reviewCmd.Flags().StringP("repo", "r", "", "Filter by repository (owner/repo)")
	reviewCmd.Flags().StringP("start-date", "s", "", "Start date for review search (YYYY-MM-DD)")
	reviewCmd.Flags().StringP("end-date", "e", "", "End date for review search (YYYY-MM-DD)")

	// Bind flags to viper for use across the app
	viper.BindPFlag("repo", reviewCmd.Flags().Lookup("repo"))
	viper.BindPFlag("start-date", reviewCmd.Flags().Lookup("start-date"))
	viper.BindPFlag("end-date", reviewCmd.Flags().Lookup("end-date"))
}
