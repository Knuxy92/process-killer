package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"process-killer/process"
)

// Styles
var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#E74C3C")).
			Padding(0, 2).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	filterCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

type Model struct {
	table       table.Model
	processes   []process.ProcessInfo
	filtered    []process.ProcessInfo
	searchInput textinput.Model
	searchMode  bool
	killTree    bool
	quitting    bool
	statusMsg   string
	statusType  string
	showConfirm bool
	confirmPID  int32
	confirmName string
	confirmAll  bool // true when confirming "kill all matching"
	ready       bool
}

// Messages
type refreshMsg struct {
	processes []process.ProcessInfo
	err       error
}

type killResultMsg struct {
	pid   int32
	name  string
	count int // number of processes killed (for kill-all)
	err   error
}

type killAllResultMsg struct {
	name  string
	count int
	errs  int
}

// Commands
func refreshProcesses() tea.Msg {
	procs, err := process.ListProcesses()
	return refreshMsg{processes: procs, err: err}
}

func killProcessCmd(pid int32, name string, killTree bool) tea.Cmd {
	return func() tea.Msg {
		err := process.KillProcess(pid, killTree)
		return killResultMsg{pid: pid, name: name, count: 1, err: err}
	}
}

func killAllMatchingCmd(procs []process.ProcessInfo, filter string) tea.Cmd {
	return func() tea.Msg {
		var matching []process.ProcessInfo
		filterLower := strings.ToLower(filter)
		for _, p := range procs {
			if strings.Contains(strings.ToLower(p.Name), filterLower) ||
				strings.Contains(strings.ToLower(p.CMDLine), filterLower) {
				matching = append(matching, p)
			}
		}

		successCount := 0
		errCount := 0
		for _, p := range matching {
			err := process.KillProcess(p.PID, false)
			if err != nil {
				errCount++
			} else {
				successCount++
			}
		}

		return killAllResultMsg{
			count: successCount,
			errs:  errCount,
		}
	}
}

// NewModel creates a new TUI model with the given table.
func NewModel(t table.Model) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search processes..."
	ti.CharLimit = 50
	ti.Width = 40

	return Model{
		table:       t,
		searchInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return refreshProcesses
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode first
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				return m, nil
			case "esc":
				m.searchMode = false
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if !m.showConfirm {
				m.quitting = true
				return m, tea.Quit
			}

		case "r":
			if !m.showConfirm {
				m.statusMsg = "Refreshing..."
				m.statusType = "info"
				return m, refreshProcesses
			}

		case "t":
			if !m.showConfirm {
				m.killTree = !m.killTree
				if m.killTree {
					m.statusMsg = "Kill tree mode: ON - will kill all child processes"
				} else {
					m.statusMsg = "Kill tree mode: OFF - will kill only the selected process"
				}
				m.statusType = "info"
				return m, nil
			}

		case "/":
			if !m.showConfirm {
				m.searchMode = true
				m.searchInput.Focus()
				m.statusMsg = "Type to search/filter processes..."
				m.statusType = "info"
				return m, nil
			}

		case "K":
			if !m.showConfirm && m.searchInput.Value() != "" {
				matching := m.getMatchingProcesses()
				if len(matching) > 0 {
					m.showConfirm = true
					m.confirmPID = 0 // Reset single PID so Enter knows it's a kill-all
					m.confirmName = m.searchInput.Value()
					m.statusMsg = fmt.Sprintf(
						"⚠️  Kill ALL %d processes matching \"%s\"? Press Enter to confirm, Esc to cancel",
						len(matching),
						m.searchInput.Value(),
					)
					m.statusType = "info"
				}
			}

		case "enter":
			if m.showConfirm {
				if m.confirmPID != 0 {
					return m, killProcessCmd(m.confirmPID, m.confirmName, m.killTree)
				}
				// Kill all matching
				return m, killAllMatchingCmd(m.processes, m.searchInput.Value())
			}

			// Show confirmation dialog for single process
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil {
				pidStr := selectedRow[0]
				pid, err := strconv.Atoi(pidStr)
				if err == nil {
					m.showConfirm = true
					m.confirmPID = int32(pid)
					m.confirmName = selectedRow[1]
					m.statusMsg = fmt.Sprintf(
						"⚠️  Kill %s (PID: %d)%s? Press Enter to confirm, Esc to cancel",
						m.confirmName,
						m.confirmPID,
						func() string {
							if m.killTree {
								return " + all children"
							}
							return ""
						}(),
					)
					m.statusType = "info"
				}
			}

		case "esc":
			if m.showConfirm {
				m.showConfirm = false
				m.statusMsg = ""
				m.statusType = ""
				return m, nil
			}

		default:
			if m.showConfirm {
				return m, nil
			}
		}

	case refreshMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
			m.statusType = "error"
			return m, nil
		}
		m.processes = msg.processes
		m.ready = true
		m.showConfirm = false
		m.applyFilter()
		return m, nil

	case killResultMsg:
		m.showConfirm = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("❌ Failed to kill %s (PID: %d): %s", msg.name, msg.pid, msg.err.Error())
			m.statusType = "error"
		} else {
			treeStr := ""
			if m.killTree {
				treeStr = " (including child processes)"
			}
			m.statusMsg = fmt.Sprintf("✅ Killed %s (PID: %d)%s", msg.name, msg.pid, treeStr)
			m.statusType = "success"
		}
		return m, refreshProcesses

	case killAllResultMsg:
		m.showConfirm = false
		if msg.errs > 0 {
			m.statusMsg = fmt.Sprintf("⚠️  Killed %d processes, %d failed (matching \"%s\")", msg.count, msg.errs, m.searchInput.Value())
			m.statusType = "error"
		} else {
			m.statusMsg = fmt.Sprintf("✅ Killed all %d processes matching \"%s\"", msg.count, m.searchInput.Value())
			m.statusType = "success"
		}
		return m, refreshProcesses
	}

	// Update table
	m.table, _ = m.table.Update(msg)
	return m, nil
}

func (m *Model) applyFilter() {
	filter := strings.TrimSpace(m.searchInput.Value())
	if filter == "" {
		// Show all processes
		rows := make([]table.Row, 0, len(m.processes))
		for _, p := range m.processes {
			rows = append(rows, table.Row{
				strconv.Itoa(int(p.PID)),
				truncateString(p.Name, 30),
				fmt.Sprintf("%.1f", p.CPU),
				fmt.Sprintf("%.1f", p.Memory),
				p.Status,
				truncateString(p.CMDLine, 50),
			})
		}
		m.table.SetRows(rows)
		m.statusMsg = fmt.Sprintf("Showing all %d processes", len(m.processes))
		m.statusType = "info"
		return
	}

	// Filter processes
	filterLower := strings.ToLower(filter)
	var filtered []process.ProcessInfo
	for _, p := range m.processes {
		if strings.Contains(strings.ToLower(p.Name), filterLower) ||
			strings.Contains(strings.ToLower(p.CMDLine), filterLower) {
			filtered = append(filtered, p)
		}
	}

	rows := make([]table.Row, 0, len(filtered))
	for _, p := range filtered {
		rows = append(rows, table.Row{
			strconv.Itoa(int(p.PID)),
			truncateString(p.Name, 30),
			fmt.Sprintf("%.1f", p.CPU),
			fmt.Sprintf("%.1f", p.Memory),
			p.Status,
			truncateString(p.CMDLine, 50),
		})
	}
	m.table.SetRows(rows)
	m.statusMsg = fmt.Sprintf("Showing %d / %d processes (filter: \"%s\")", len(filtered), len(m.processes), filter)
	m.statusType = "info"
}

func (m Model) getMatchingProcesses() []process.ProcessInfo {
	filter := strings.TrimSpace(m.searchInput.Value())
	if filter == "" {
		return nil
	}
	filterLower := strings.ToLower(filter)
	var matching []process.ProcessInfo
	for _, p := range m.processes {
		if strings.Contains(strings.ToLower(p.Name), filterLower) ||
			strings.Contains(strings.ToLower(p.CMDLine), filterLower) {
			matching = append(matching, p)
		}
	}
	return matching
}

func (m Model) View() string {
	if !m.ready {
		return "Loading processes..."
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(" 🔪 Process Killer "))
	b.WriteString("\n")

	// Search bar
	if m.searchMode {
		b.WriteString("\n")
		b.WriteString(searchStyle.Render("🔍 Search: "))
		b.WriteString(m.searchInput.View())
		b.WriteString("\n\n")
	} else if m.searchInput.Value() != "" {
		b.WriteString("\n")
		b.WriteString(searchStyle.Render(fmt.Sprintf("🔍 Filter: \"%s\"", m.searchInput.Value())))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
	}

	// Table
	b.WriteString(baseStyle.Render(m.table.View()))
	b.WriteString("\n\n")

	// Status bar
	if m.statusMsg != "" {
		var styled string
		switch m.statusType {
		case "error":
			styled = errorStyle.Render(m.statusMsg)
		case "success":
			styled = successStyle.Render(m.statusMsg)
		case "info":
			styled = statusStyle.Render(m.statusMsg)
		default:
			styled = m.statusMsg
		}
		b.WriteString(styled)
		b.WriteString("\n\n")
	}

	// Help
	helpText := "↑/↓ navigate • Enter select • / search • K kill all matching • t toggle kill-tree • r refresh • q quit"
	if m.searchMode {
		helpText = "Type to search • Enter done • Esc clear"
	}
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
