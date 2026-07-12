package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"process-killer/tui"
)

func main() {
	// Define table columns
	columns := []table.Column{
		{Title: "PID", Width: 8},
		{Title: "Name", Width: 32},
		{Title: "CPU%", Width: 7},
		{Title: "MEM%", Width: 7},
		{Title: "Status", Width: 10},
		{Title: "Command", Width: 52},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("196")).
		Bold(true)
	t.SetStyles(s)

	m := tui.NewModel(t)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
