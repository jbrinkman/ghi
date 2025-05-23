/*
Copyright © 2024 Joe Brinkman <joe.brinkman@improving.com>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v69/github"
	"github.com/jbrinkman/ghi/pkg/clients"
	gh "github.com/jbrinkman/ghi/pkg/github"
	"github.com/jbrinkman/ghi/pkg/logger"
	"github.com/jbrinkman/ghi/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// prCmd represents the pullrequest command
var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "List pull requests from a specified GitHub repository",
	Long: `The 'pr' command retrieves and lists pull requests from a specified GitHub repository.
You can filter the pull requests by author using the --author option. Multiple --author options
can be used to provide a list of author filters. The command outputs the number, title, author,
state, and URL of each pull request.`,
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
		viper.BindPFlag("author", cmd.Flags().Lookup("author"))
		viper.BindPFlag("state", cmd.Flags().Lookup("state"))
		viper.BindPFlag("reviewer", cmd.Flags().Lookup("reviewer"))
		viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
		viper.BindPFlag("draft", cmd.Flags().Lookup("draft"))

		repo := viper.GetString("repo")
		if repo == "" {
			log.Fatal("The --repo flag is required")
		}

		debug := viper.GetBool("debug")
		// Debug logging is handled by the root command's PersistentPreRun

		authors := viper.GetStringSlice("author")
		state := viper.GetString("state")
		reviewers := viper.GetStringSlice("reviewer")
		draftOption := viper.GetString("draft")

		// Convert authors and reviewers to lowercase for case-insensitive comparison
		for i, author := range authors {
			authors[i] = strings.ToLower(author)
		}
		for i, reviewer := range reviewers {
			reviewers[i] = strings.ToLower(reviewer)
		}

		// Split the repo into owner and repo name
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			log.Fatal("Invalid repository format. Use 'owner/repo'")
		}
		owner, repoName := parts[0], parts[1]

		if debug {
			logger.Debug("Command arguments: %v", args)
			logger.Debug("Repository: %s/%s", owner, repoName)
			logger.Debug("Authors filter: %v", authors)
			logger.Debug("State filter: %s", state)
			logger.Debug("Reviewers filter: %v", reviewers)
			logger.Debug("Draft option: %s", draftOption)
		}

		// Create a new Github client with cache control
		ctx := context.Background()
		client, err := clients.NewGitHubClient()
		if err != nil {
			log.Fatalf("Failed to create GitHub client: %v", err)
		}

		// Construct the search query
		query := fmt.Sprintf("repo:%s/%s", owner, repoName)
		if state != "" && state != "all" {
			query += fmt.Sprintf(" state:%s", state)
		}
		for _, author := range authors {
			query += fmt.Sprintf(" author:%s", author)
		}
		query += " type:pr" // Ensure only pull requests are returned

		if debug {
			logger.Debug("Search query: %s", query)
		}

		// Search pull requests with retry on rate limit
		var result *github.IssuesSearchResult
		scanPRs := func() (*github.IssuesSearchResult, error) {
			searchOpts := &github.SearchOptions{}
			for attempts := 0; attempts < 3; attempts++ {
				var err error
				result, _, err = client.Search.Issues(ctx, query, searchOpts)
				if err != nil {
					if _, ok := err.(*github.RateLimitError); ok {
						if attempts < 2 {
							// On rate limit, wait and retry
							logger.Debug("Hit rate limit, waiting 5 seconds before retry...")
							time.Sleep(5 * time.Second)
							continue
						}
						return nil, fmt.Errorf("GitHub API rate limit exceeded. Try setting GHI_GITHUB_TOKEN environment variable")
					}
					return nil, fmt.Errorf("error searching pull requests: %w", err)
				}
				return result, nil
			}
			return nil, fmt.Errorf("max retries exceeded")
		}

		// Show spinner while fetching PRs
		logger.Debug("Starting to fetch pull requests with query: %s", query)
		result, err = ui.WithSpinner(ctx, "Fetching pull requests", scanPRs)
		if err != nil {
			logger.Debug("Error fetching pull requests: %v", err)
			log.Fatal(err)
		}
		logger.Debug("Found %d issues from search", len(result.Issues))

		if debug {
			logger.Debug("Found %d pull requests", len(result.Issues))
		}

		// Process PRs with a spinner
		processPRs := func() ([]*gh.PullRequestData, error) {
			logger.Debug("Creating new PR collection for %s/%s", owner, repoName)
			collection := gh.NewPRCollection(ctx, client, owner, repoName, debug)
			collection.WithDraftOption(draftOption)

			// Process the data in a pipeline
			logger.Debug("Fetching issues (count: %d)", len(result.Issues))
			collection.FetchIssues(result.Issues)
			logger.Debug("Enriching with pull requests")
			collection.EnrichWithPullRequests()
			logger.Debug("Enriching with reviews for reviewers: %v", reviewers)
			collection.EnrichWithReviews(reviewers)
			logger.Debug("Filtering drafts with option: %s", draftOption)
			collection.FilterDrafts()

			logger.Debug("Processing complete. Found %d PRs after filtering", len(collection.Items))
			for i, item := range collection.Items {
				if i >= 5 { // Only show first 5 items
					logger.Debug("  ... and %d more items", len(collection.Items)-5)
					break
				}
				if item != nil && item.Issue != nil && item.Issue.Number != nil {
					logger.Debug("  PR #%d: %s", *item.Issue.Number, *item.Issue.Title)
				}
			}

			return collection.Items, nil
		}

		// Show spinner while processing PRs
		prItems, err := ui.WithSpinner(ctx, "Processing pull requests", processPRs)
		if err != nil {
			log.Fatal(err)
		}

		// Create and show the interactive table
		prTable := ui.NewPRTable(prItems)
		p := tea.NewProgram(prTable, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running PR table: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(prCmd)

	// Define flags
	prCmd.Flags().StringP("repo", "r", "", "The name of the Github repository (owner/repo)")
	prCmd.Flags().StringArrayP("author", "A", []string{}, "Filter pull requests by author")
	prCmd.Flags().StringP("state", "s", "all", "Filter pull requests by state (ALL, OPEN, CLOSED)")
	prCmd.Flags().StringArrayP("reviewer", "R", []string{}, "Highlight pull requests by reviewer")
	prCmd.Flags().StringP("config", "c", "", "Path to the configuration file")
	prCmd.Flags().StringP("draft", "D", "hide", "Control draft PR display (show, hide)")
}

func prettyPrint(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
