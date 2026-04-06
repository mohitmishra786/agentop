package ui

import (
	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	sessions   []*aggregator.SessionStats
	termWidth  int
	termHeight int
	spinner    spinner.Model
	loading    bool
	err        error
}

func NewModel(sessions []*aggregator.SessionStats) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleGreen

	return Model{
		sessions: sessions,
		spinner:  s,
		loading:  false,
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

	return RenderToday(m.sessions, m.termWidth)
}
