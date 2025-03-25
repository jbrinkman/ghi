// Package github provides display utilities for GitHub data.
package github

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/jbrinkman/ghi/pkg/logger"
	"github.com/jedib0t/go-pretty/v6/table"
)

// DisplayOptions contains configuration for how to display PR data
type DisplayOptions struct {
	ShowDraft    bool
	ShowReviewer bool
	Debug        bool
}

// PRDisplay handles the display of pull request data
type PRDisplay struct {
	Collection *PRCollection
	Options    DisplayOptions
}

// NewPRDisplay creates a new display for PR data
func NewPRDisplay(collection *PRCollection) *PRDisplay {
	return &PRDisplay{
		Collection: collection,
		Options: DisplayOptions{
			ShowDraft:    true, // Always show draft status
			ShowReviewer: false,
			Debug:        collection.Debug,
		},
	}
}

// WithReviewers configures the display to show reviewer columns
func (d *PRDisplay) WithReviewers(hasReviewers bool) *PRDisplay {
	d.Options.ShowReviewer = hasReviewers
	return d
}

// RenderTable displays the PR collection as a formatted table
func (d *PRDisplay) RenderTable() {
	items := d.Collection.Items
	owner := d.Collection.Owner
	repo := d.Collection.Repo

	fmt.Println("=====================================")
	fmt.Printf("Pull requests for %s/%s\n", owner, repo)
	fmt.Printf("Count: %d\n", len(items))
	fmt.Println("=====================================")
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Configure table style for better readability
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateRows = false

	// Set up headers
	d.appendTableHeader(t)

	// Add table rows
	for _, prData := range items {
		d.appendTableRow(t, prData)
	}

	t.Render()

	if d.Options.Debug {
		logger.Debug("Pull request table rendered with %d rows", len(items))
	}
}

// appendTableHeader adds the appropriate header row to the table based on options
func (d *PRDisplay) appendTableHeader(t table.Writer) {
	var header table.Row

	// Always show these base columns
	header = append(header, "NUMBER", "TITLE", "AUTHOR", "STATE", "REVIEWS")

	// Show draft status column only when we're showing drafts
	if d.Collection.DraftOption == "show" {
		header = append(header, "DRAFT")
	}

	// Show reviewer status when reviewers are specified
	if d.Options.ShowReviewer {
		header = append(header, "REVIEWER")
	}

	// Always show approvals
	header = append(header, "APPROVALS")

	t.AppendHeader(header)
}

// appendTableRow formats and adds a PR data row to the table
func (d *PRDisplay) appendTableRow(t table.Writer, prData *PullRequestData) {
	if prData == nil || prData.Issue == nil {
		return
	}

	// Always included columns
	row := table.Row{
		formatPRNumber(prData),
		formatTitle(prData, d.Options.ShowDraft),
		getUserLogin(prData.Issue.User),
		getState(prData.Issue),
		len(prData.Reviews), // Show total review count
	}

	// Show draft status column only when we're showing drafts
	if d.Collection.DraftOption == "show" {
		draftStatus := "[ ]"
		if prData.IsDraft {
			draftStatus = "[X]"
		}
		row = append(row, draftStatus)
	}

	// Show reviewer status when reviewers are specified
	if d.Options.ShowReviewer {
		row = append(row, prData.ReviewerStatus)
	}

	// Always show approvals
	row = append(row, prData.ApprovalCount)

	t.AppendRow(row)
}

// Helper functions for formatting

func formatPRNumber(prData *PullRequestData) string {
	if prData.Issue == nil || prData.Issue.Number == nil {
		return "N/A"
	}

	prNumber := fmt.Sprintf("%d", *prData.Issue.Number)
	createdAt := prData.Issue.CreatedAt

	if createdAt == nil {
		return prNumber
	}

	timestamp := createdAt.GetTime()
	if timestamp == nil {
		return prNumber
	}

	daysOld := time.Since(*timestamp).Hours() / 24

	// Color priority: age over 30 days is always red, then drafts are gray, new PRs are green
	if daysOld > 30 {
		return fmt.Sprintf("\033[31m%d\033[0m", *prData.Issue.Number) // Red for old PRs
	} else if prData.IsDraft {
		return fmt.Sprintf("\033[90m%d\033[0m", *prData.Issue.Number) // Mid-gray for drafts
	} else if daysOld <= 1 {
		return fmt.Sprintf("\033[32m%d\033[0m", *prData.Issue.Number) // Green for new PRs
	}

	return prNumber
}

func formatTitle(prData *PullRequestData, showDraft bool) string {
	if prData.Issue == nil || prData.Issue.Title == nil {
		return "N/A"
	}

	title := *prData.Issue.Title

	// Add DRAFT: prefix to title if it's a draft PR and we're showing drafts
	if prData.IsDraft && showDraft {
		title = "DRAFT: " + title
	}

	// Truncate long titles
	if len(title) > 25 {
		title = title[:22] + "..."
	}

	return title
}

func getUserLogin(user *github.User) string {
	if user != nil && user.Login != nil {
		return *user.Login
	}
	return "unknown"
}

func getState(issue *github.Issue) string {
	if issue != nil && issue.State != nil {
		return *issue.State
	}
	return "unknown"
}
