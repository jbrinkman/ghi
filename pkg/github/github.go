// Package github provides a fluent API for interacting with GitHub data.
// It allows for progressive enrichment of data structures by chaining method calls.
package github

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/jbrinkman/ghi/pkg/logger"
)

// PullRequestData represents a consolidated view of a GitHub pull request
// with all related information needed for display.
type PullRequestData struct {
	Issue           *github.Issue
	PullRequest     *github.PullRequest
	Reviews         []*github.PullRequestReview
	Authors         []string
	Reviewers       []string
	UniqueReviewers map[string]struct{}
	ApprovalCount   int
	ReviewerStatus  string
	IsDraft         bool
	DraftStatus     string
}

// PRCollection holds a collection of pull request data and context for operations
type PRCollection struct {
	Items       []*PullRequestData
	Client      *github.Client
	Owner       string
	Repo        string
	Context     context.Context
	Debug       bool
	DraftOption string
}

// NewPRCollection creates a new PRCollection with the given client and context
func NewPRCollection(ctx context.Context, client *github.Client, owner, repo string, debug bool) *PRCollection {
	return &PRCollection{
		Items:   make([]*PullRequestData, 0),
		Client:  client,
		Owner:   owner,
		Repo:    repo,
		Context: ctx,
		Debug:   debug,
	}
}

// WithDraftOption sets the draft display option for the collection
func (c *PRCollection) WithDraftOption(option string) *PRCollection {
	c.DraftOption = option
	return c
}

// FetchIssues retrieves issues from GitHub and initializes the PR collection
func (c *PRCollection) FetchIssues(issues []*github.Issue) *PRCollection {
	for _, issue := range issues {
		prData := &PullRequestData{
			Issue:           issue,
			UniqueReviewers: make(map[string]struct{}),
		}
		c.Items = append(c.Items, prData)
	}

	if c.Debug {
		logger.Debug("Initialized %d pull request data objects", len(c.Items))
	}

	return c
}

// handleRateLimit attempts to handle rate limit errors with retries
func (c *PRCollection) handleRateLimit(err error) bool {
	if _, ok := err.(*github.RateLimitError); ok {
		// If we hit rate limit, wait 5 seconds and try again
		if c.Debug {
			logger.Debug("Hit rate limit, waiting 5 seconds before retry...")
		}
		time.Sleep(5 * time.Second)
		return true
	}
	return false
}

// EnrichWithPullRequests retrieves and attaches pull request data for each issue
func (c *PRCollection) EnrichWithPullRequests() *PRCollection {
	for i, prData := range c.Items {
		if c.Debug {
			logger.Debug("Fetching PR details for #%d (%d of %d)",
				*prData.Issue.Number, i+1, len(c.Items))
		}

		// Try up to 3 times if we hit rate limits
		var pr *github.PullRequest
		var err error
		for attempts := 0; attempts < 3; attempts++ {
			pr, _, err = c.Client.PullRequests.Get(c.Context, c.Owner, c.Repo, *prData.Issue.Number)
			if err != nil {
				if attempts < 2 && c.handleRateLimit(err) {
					continue
				}
				if c.Debug {
					logger.Debug("Error fetching PR details for #%d: %v", *prData.Issue.Number, err)
				}
				break
			}
			break
		}

		if err != nil {
			continue
		}

		prData.PullRequest = pr
		prData.IsDraft = pr.GetDraft()

		if prData.IsDraft {
			prData.DraftStatus = "[X]"
			if c.Debug {
				logger.Debug("PR #%d is a draft", *prData.Issue.Number)
			}
		} else {
			prData.DraftStatus = "[ ]"
		}
	}

	return c
}

// EnrichWithReviews retrieves and processes review information for each PR
func (c *PRCollection) EnrichWithReviews(reviewers []string) *PRCollection {
	// Convert all reviewers to lowercase for case-insensitive comparison
	lowercaseReviewers := make([]string, len(reviewers))
	for i, r := range reviewers {
		lowercaseReviewers[i] = strings.ToLower(r)
	}

	if c.Debug {
		logger.Debug("Starting review enrichment for %d PRs with %d reviewers: %v",
			len(c.Items), len(reviewers), reviewers)
	}

	for i, prData := range c.Items {
		// Always initialize reviewer status to [ ] for all PRs
		prData.ReviewerStatus = "[ ]"

		if c.Debug {
			logger.Debug("Fetching reviews for PR #%d (%d of %d)",
				*prData.Issue.Number, i+1, len(c.Items))
		}

		// Try up to 3 times if we hit rate limits
		var reviews []*github.PullRequestReview
		var err error
		for attempts := 0; attempts < 3; attempts++ {
			reviews, _, err = c.Client.PullRequests.ListReviews(
				c.Context, c.Owner, c.Repo, *prData.Issue.Number, nil)
			if err != nil {
				if attempts < 2 && c.handleRateLimit(err) {
					continue
				}
				if c.Debug {
					logger.Debug("Error fetching reviews for PR #%d: %v", *prData.Issue.Number, err)
				}
				break
			}
			break
		}

		if err != nil {
			continue
		}

		prData.Reviews = reviews

		if c.Debug {
			logger.Debug("PR #%d has %d reviews", *prData.Issue.Number, len(reviews))
		}

		// Always process review counts, regardless of whether reviewers were specified
		prAuthor := strings.ToLower(getPRAuthor(prData))
		reviewerFound := false

		for _, review := range reviews {
			reviewer := strings.ToLower(getReviewerLogin(review))
			reviewState := "none"
			if review.State != nil {
				reviewState = *review.State
			}

			if c.Debug {
				logger.Debug("Processing review by %s with state: %s", reviewer, reviewState)
			}

			// Check if this PR has been reviewed by one of the specified reviewers
			if len(lowercaseReviewers) > 0 &&
				contains(lowercaseReviewers, reviewer) &&
				isApprovedOrCommented(review) {
				reviewerFound = true
				prData.ReviewerStatus = "[X]"
				if c.Debug {
					logger.Debug("PR #%d has been reviewed by specified reviewer: %s",
						*prData.Issue.Number, reviewer)
				}
			}

			// Count unique reviewers (excluding the PR author)
			if reviewer != "" && reviewer != prAuthor && isApprovedOrCommented(review) {
				prData.UniqueReviewers[reviewer] = struct{}{}
				if c.Debug {
					logger.Debug("Added %s to unique reviewers for PR #%d",
						reviewer, *prData.Issue.Number)
				}
			}

			// Count approvals
			if isApproved(review) {
				prData.ApprovalCount++
				if c.Debug {
					logger.Debug("Incremented approval count for PR #%d (now: %d)",
						*prData.Issue.Number, prData.ApprovalCount)
				}
			}
		}

		if c.Debug {
			logger.Debug("PR #%d processing complete: %d unique reviewers, %d approvals, reviewer found: %v",
				*prData.Issue.Number, len(prData.UniqueReviewers), prData.ApprovalCount, reviewerFound)
		}
	}

	return c
}

// FilterDrafts removes draft PRs from the collection if draftOption is "hide"
func (c *PRCollection) FilterDrafts() *PRCollection {
	if c.Debug {
		logger.Debug("FilterDrafts called with draftOption: %s", c.DraftOption)
	}

	if c.DraftOption != "hide" {
		if c.Debug {
			logger.Debug("Not filtering drafts because draftOption is not 'hide'")
		}
		return c
	}

	if c.Debug {
		logger.Debug("Starting to filter draft PRs, current count: %d", len(c.Items))

		// Count how many are drafts
		draftCount := 0
		for _, prData := range c.Items {
			if prData.IsDraft {
				draftCount++
			}
		}
		logger.Debug("Found %d draft PRs to filter out", draftCount)
	}

	filtered := make([]*PullRequestData, 0)
	for _, prData := range c.Items {
		if !prData.IsDraft {
			filtered = append(filtered, prData)
		} else if c.Debug {
			logger.Debug("Filtering out draft PR #%d (isDraft=%v)",
				*prData.Issue.Number, prData.IsDraft)
		}
	}

	if c.Debug {
		logger.Debug("After filtering, PR count reduced from %d to %d",
			len(c.Items), len(filtered))
	}

	c.Items = filtered
	return c
}

// GetItems returns the final collection of PR data
func (c *PRCollection) GetItems() []*PullRequestData {
	return c.Items
}

// Helper functions

func getPRAuthor(prData *PullRequestData) string {
	if prData.Issue != nil && prData.Issue.User != nil && prData.Issue.User.Login != nil {
		return *prData.Issue.User.Login
	}
	return ""
}

func getReviewerLogin(review *github.PullRequestReview) string {
	if review != nil && review.User != nil && review.User.Login != nil {
		return *review.User.Login
	}
	return ""
}

// isApprovedOrCommented safely checks if a review has APPROVED or COMMENTED state
func isApprovedOrCommented(review *github.PullRequestReview) bool {
	if review == nil || review.State == nil {
		return false
	}
	return *review.State == "APPROVED" || *review.State == "COMMENTED"
}

// isApproved safely checks if a review has APPROVED state
func isApproved(review *github.PullRequestReview) bool {
	if review == nil || review.State == nil {
		return false
	}
	return *review.State == "APPROVED"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
