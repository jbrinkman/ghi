/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
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
		repo, _ := cmd.Flags().GetString("repo")
		if repo == "" {
			log.Fatal("The --repo flag is required")
		}

		authors, _ := cmd.Flags().GetStringArray("author")

		// Convert authors to lowercase for case-insensitive comparison
		for i, author := range authors {
			authors[i] = strings.ToLower(author)
		}

		// Split the repo into owner and repo name
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			log.Fatal("Invalid repository format. Use 'owner/repo'")
		}
		owner, repoName := parts[0], parts[1]

		// Create a new Github client
		ctx := context.Background()
		client := github.NewClient(nil)

		// List pull requests
		prs, _, err := client.PullRequests.List(ctx, owner, repoName, nil)
		if err != nil {
			log.Fatalf("Error fetching pull requests: %v", err)
		}

		fmt.Printf("Pull requests for %s/%s\n", owner, repoName)
		fmt.Println("=====================================")

		for _, pr := range prs {
			if len(authors) > 0 {
				author := strings.ToLower(*pr.User.Login)
				if !contains(authors, author) {
					continue
				}
			}
			fmt.Printf("#%d - %s\n", *pr.Number, *pr.Title)
			fmt.Printf("Author: %s\n", *pr.User.Login)
			fmt.Printf("State: %s\n", *pr.State)
			fmt.Printf("URL: %s\n", *pr.HTMLURL)
			fmt.Println("=====================================")
		}
	},
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(prCmd)

	// Define the --repo flag
	prCmd.Flags().StringP("repo", "r", "", "The name of the Github repository (owner/repo)")
	prCmd.MarkFlagRequired("repo")

	// Define the --author flag
	prCmd.Flags().StringArrayP("author", "a", []string{}, "Filter pull requests by author")
}
