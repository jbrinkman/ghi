package ui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoaderModel struct {
	spinner  spinner.Model
	message  string
	complete bool
}

func NewLoader(message string) *LoaderModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &LoaderModel{
		spinner:  s,
		message:  message,
		complete: false,
	}
}

func (m LoaderModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

func (m LoaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m LoaderModel) View() string {
	if m.complete {
		return ""
	}
	return fmt.Sprintf("%s %s...", m.spinner.View(), m.message)
}

// WithSpinner runs the provided function while showing a loading spinner.
func WithSpinner(ctx context.Context, message string, fn func() error) error {
	loader := NewLoader(message)
	p := tea.NewProgram(loader)

	// Start the spinner in a goroutine
	done := make(chan error, 1)
	go func() {
		err := fn()
		loader.complete = true
		p.Send(tea.Quit())
		done <- err
	}()

	// Run the TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run spinner: %w", err)
	}

	// Return the result from the function
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
