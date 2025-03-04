// Package db provides functionality for tracking code reviews using Turso/LibSQL.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

const (
	// ReviewsTableName is the name of the table storing review data
	ReviewsTableName = "reviews"
)

// Review represents a code review entry in the database
type Review struct {
	ID        int64
	Repo      string
	PRNumber  int
	Reviewer  string
	Timestamp time.Time
}

// Client handles database operations for review tracking
type Client struct {
	db *sql.DB
}

// NewClient creates a new database client using environment variables for configuration
func NewClient() (*Client, error) {
	dbURL := os.Getenv("GHI_DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("GHI_DB_URL environment variable not set")
	}

	authToken := os.Getenv("GHI_AUTH_TOKEN")

	// Create the connection string with auth token if available
	connStr := dbURL
	if authToken != "" {
		connStr = fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	}

	// Open a connection to the database
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database (open): %w", err)
	}

	// Test connection
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database (ping): %w", err)
	}

	return &Client{db: db}, nil
}

// InitSchema ensures the database schema exists
func (c *Client) InitSchema(ctx context.Context) error {
	// Create reviews table if it doesn't exist
	_, err := c.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo TEXT NOT NULL,
			pr_number INTEGER NOT NULL,
			reviewer TEXT NOT NULL,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(repo, pr_number, reviewer, timestamp)
		)
	`)

	return err
}

// LogReview records a new code review in the database
func (c *Client) LogReview(ctx context.Context, repo string, prNumber int, reviewer string) error {
	_, err := c.db.ExecContext(ctx,
		"INSERT INTO reviews (repo, pr_number, reviewer) VALUES (?, ?, ?)",
		repo, prNumber, reviewer)

	if err != nil {
		return fmt.Errorf("failed to log review: %w", err)
	}

	return nil
}

// parseTimestamp attempts to parse a timestamp string using multiple formats
func parseTimestamp(timestamp string) (time.Time, error) {
	// Try different time formats, from most specific to least specific
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z", // ISO8601 without timezone
		"2006-01-02 15:04:05",  // Simple date-time format
		"2006-01-02",           // Just the date
	}

	var err error
	var t time.Time

	for _, format := range formats {
		t, err = time.Parse(format, timestamp)
		if err == nil {
			return t, nil
		}
	}

	// If we get here, none of the formats worked
	return time.Time{}, fmt.Errorf("failed to parse timestamp '%s': %w", timestamp, err)
}

// GetReviews retrieves reviews for a specific PR
func (c *Client) GetReviews(ctx context.Context, repo string, prNumber int) ([]Review, error) {
	rows, err := c.db.QueryContext(ctx,
		"SELECT id, repo, pr_number, reviewer, timestamp FROM reviews WHERE repo = ? AND pr_number = ? ORDER BY timestamp DESC",
		repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var timestamp string

		err := rows.Scan(&review.ID, &review.Repo, &review.PRNumber, &review.Reviewer, &timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review row: %w", err)
		}

		// Parse timestamp with the flexible parser
		t, err := parseTimestamp(timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		review.Timestamp = t

		reviews = append(reviews, review)
	}

	return reviews, nil
}

// GetReviewsByDateRange retrieves reviews within the specified date range
func (c *Client) GetReviewsByDateRange(ctx context.Context, repo string, startDate, endDate time.Time) ([]Review, error) {
	// Construct query based on whether we have a repo filter
	var query string
	var args []interface{}

	if repo != "" {
		query = `SELECT id, repo, pr_number, reviewer, timestamp 
				FROM reviews 
				WHERE repo = ? AND timestamp >= ? AND timestamp <= ? 
				ORDER BY timestamp DESC`
		args = []interface{}{repo, startDate.Format("2006-01-02"), endDate.Format("2006-01-02 23:59:59")}
	} else {
		query = `SELECT id, repo, pr_number, reviewer, timestamp 
				FROM reviews 
				WHERE timestamp >= ? AND timestamp <= ? 
				ORDER BY timestamp DESC`
		args = []interface{}{startDate.Format("2006-01-02"), endDate.Format("2006-01-02 23:59:59")}
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews by date range: %w", err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		var timestamp string

		err := rows.Scan(&review.ID, &review.Repo, &review.PRNumber, &review.Reviewer, &timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review row: %w", err)
		}

		// Parse timestamp with the flexible parser
		t, err := parseTimestamp(timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		review.Timestamp = t

		reviews = append(reviews, review)
	}

	return reviews, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}
