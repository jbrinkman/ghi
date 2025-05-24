package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jbrinkman/ghi/pkg/logger"
	gh "github.com/jbrinkman/ghi/pkg/github"
)

type PRTableModel struct {
	table  table.Model
	prData  []*gh.PullRequestData
	loading bool
	err     error
}

// createTableRows converts PR data to table rows
func createTableRows(prData []*gh.PullRequestData) []table.Row {
	var rows []table.Row
	for _, pr := range prData {
		if pr == nil || pr.Issue == nil || pr.Issue.Number == nil {
			continue
		}

		author := "unknown"
		if pr.Issue.User != nil && pr.Issue.User.Login != nil {
			author = *pr.Issue.User.Login
		}

		reviewStatus := ""
		if pr.ReviewerStatus != "" {
			reviewStatus = pr.ReviewerStatus
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("#%d", *pr.Issue.Number),
			truncateString(*pr.Issue.Title, 35),
			truncateString(author, 12),
			*pr.Issue.State,
			reviewStatus,
		})
	}
	return rows
}

// NewPRTable creates a new Bubble Tea model for displaying PRs in a table
func NewPRTable(prData []*gh.PullRequestData) *PRTableModel {
	// Debug logging
	logger.Debug("Creating new PR table with %d items", len(prData))
	for i, pr := range prData {
		if pr == nil || pr.Issue == nil || pr.Issue.Number == nil {
			logger.Debug("  PR <nil or invalid>")
		} else {
			logger.Debug("  PR #%d: %s", *pr.Issue.Number, *pr.Issue.Title)
		}
		if i >= 4 { // Only show first 5 items to avoid log spam
			logger.Debug("  ... and %d more items", len(prData)-5)
			break
		}
	}

	columns := []table.Column{
		{Title: "#", Width: 5},
		{Title: "Title", Width: 40},
		{Title: "Author", Width: 15},
		{Title: "State", Width: 8},
		{Title: "Reviews", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(createTableRows(prData)),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &PRTableModel{
		table:   t,
		prData:  prData,
		loading: false,
	}
}

// Init initializes the table model
func (m *PRTableModel) Init() tea.Cmd {
	logger.Debug("PRTableModel.Init() called")
	if m.table.Rows() != nil {
		logger.Debug("Table has %d rows", len(m.table.Rows()))
	} else {
		logger.Debug("Table has no rows")
	}
	return nil
}

func (m *PRTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the table
func (m *PRTableModel) View() string {
	if m.loading {
		return "Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	logger.Debug("Rendering table with %d rows", len(m.table.Rows()))
	if len(m.table.Rows()) == 0 {
		return "No pull requests found"
	}
	var b strings.Builder
	b.WriteString("\n" + m.table.View() + "\n")
	b.WriteString("↑/↓: Navigate • q: Quit\n")
	return b.String()
}

// UpdatePRs updates the table with new PR data
func (m *PRTableModel) UpdatePRs(prData []*gh.PullRequestData) {
	m.prData = prData
	m.loading = false

	// Update the table with new data
	m.table.SetRows(createTableRows(prData))
}

// truncateString shortens a string to the specified length and adds "..." if truncated
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// formatDaysAgo formats a time.Time as "X days ago"
func formatDaysAgo(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	days := int(time.Since(*t).Hours() / 24)
	if days == 0 {
		return "today"
	} else if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
