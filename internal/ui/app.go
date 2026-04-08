package ui

import (
	"fmt"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	sessions      []*aggregator.SessionStats
	termWidth     int
	termHeight    int
	spinner       spinner.Model
	err           error
	showHelp      bool
	selectedIndex int
	scrollOffset  int
	sortBy        string
	filterText    string
}

func NewModel(sessions []*aggregator.SessionStats) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleGreen

	return Model{
		sessions:      sessions,
		spinner:       s,
		showHelp:      false,
		selectedIndex: 0,
		scrollOffset:  0,
		sortBy:        "time",
		filterText:    "",
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				if m.selectedIndex < m.scrollOffset {
					m.scrollOffset = m.selectedIndex
				}
			}
			return m, nil
		case "down", "j":
			if m.selectedIndex < len(m.sessions)-1 {
				m.selectedIndex++
				maxVisible := m.maxVisibleSessions()
				if m.selectedIndex >= m.scrollOffset+maxVisible {
					m.scrollOffset = m.selectedIndex - maxVisible + 1
				}
			}
			return m, nil
		case "s":
			m.cycleSort()
			return m, nil
		case "r":
			m.scrollOffset = 0
			m.selectedIndex = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return StyleRed.Render("Error: " + m.err.Error())
	}
	if m.termWidth == 0 {
		m.termWidth = 120
	}

	if m.showHelp {
		return m.renderHelp()
	}

	content := RenderToday(m.sessions, m.termWidth)
	return content + m.renderStatusBar()
}

func (m Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Border)).
		Padding(1, 2).
		Width(60)

	content := StyleBold.Render("Keyboard Shortcuts\n\n") +
		StyleDim.Render("Navigation:\n") +
		"  ↑/k    Move up\n" +
		"  ↓/j    Move down\n" +
		"  r      Reset scroll\n\n" +
		StyleDim.Render("Actions:\n") +
		"  s      Cycle sort (time/cost/tokens/cache)\n" +
		"  ?      Toggle this help\n" +
		"  q      Quit\n\n" +
		StyleDim.Render("Current sort: ") +
		StyleGreen.Render(m.sortBy)

	return helpStyle.Render(content)
}

func (m Model) renderStatusBar() string {
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextDim)).
		Background(lipgloss.Color("#222230")).
		Padding(0, 1)

	info := fmt.Sprintf(" %d sessions | Sort: %s | ↑↓ navigate | ? help | q quit",
		len(m.sessions), m.sortBy)

	return "\n" + statusBar.Render(info)
}

func (m Model) maxVisibleSessions() int {
	return max(1, m.termHeight/3-2)
}

func (m *Model) cycleSort() {
	sorts := []string{"time", "cost", "tokens", "cache"}
	for i, s := range sorts {
		if s == m.sortBy {
			m.sortBy = sorts[(i+1)%len(sorts)]
			return
		}
	}
	m.sortBy = sorts[0]
}
