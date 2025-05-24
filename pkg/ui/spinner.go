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
// The function can return a value of any type and an error.
func WithSpinner[T any](ctx context.Context, message string, fn func() (T, error)) (T, error) {
	loader := NewLoader(message)
	p := tea.NewProgram(loader)

	// Start the spinner in a goroutine
	type result struct {
		value T
		err   error
	}
	done := make(chan result, 1)
	go func() {
		val, err := fn()
		loader.complete = true
		p.Send(tea.Quit())
		done <- result{value: val, err: err}
	}()

	// Run the TUI
	if _, err := p.Run(); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to run spinner: %w", err)
	}

	// Return the result from the function
	select {
	case res := <-done:
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}
