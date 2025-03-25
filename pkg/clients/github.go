// Package clients provides HTTP client configurations for external services
package clients

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

// NewGitHubClient creates a new GitHub client with custom configuration
// to prevent caching issues and ensure fresh data on each request.
// It will use GHI_GITHUB_TOKEN environment variable for authentication if available.
func NewGitHubClient() (*github.Client, error) {
	// Check for GitHub token
	token := os.Getenv("GHI_GITHUB_TOKEN")
	var httpClient *http.Client

	if token != "" {
		// Create authenticated client with OAuth2
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(context.Background(), ts)
	} else {
		// Create unauthenticated client with custom transport
		transport := &http.Transport{
			DisableKeepAlives:     true,
			IdleConnTimeout:       1 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		}

		httpClient = &http.Client{
			Transport: transport,
			Timeout:   1 * time.Minute,
		}

		// Warn about rate limiting
		fmt.Fprintln(os.Stderr, "Warning: No GitHub token found. Requests will be rate limited to 60 per hour.")
		fmt.Fprintln(os.Stderr, "Set GHI_GITHUB_TOKEN environment variable to increase rate limit to 5000 per hour.")
	}

	// Create GitHub client
	return github.NewClient(httpClient), nil
}
