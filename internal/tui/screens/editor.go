package screens

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Editor struct {
	config   *config.Manager
	bus      *tui.EventBus
	textarea textarea.Model
	width    int
	height   int
}

func NewEditor(cfg *config.Manager, bus *tui.EventBus) *Editor {
	ta := textarea.New()
	ta.Placeholder = "Enter SQL statement..."
	ta.SetWidth(80)
	ta.SetHeight(20)
	return &Editor{config: cfg, bus: bus, textarea: ta}
}

func (m *Editor) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 6)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return tui.NavigateToMsg{Screen: tui.BrowserScreen} }
		case "ctrl+e":
			return m, nil

		}
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m *Editor) View() string {
	return tui.AppStyle.Render(
		tui.TitleStyle.Render("SQL Editor") + "\n\n" +
			m.textarea.View() + "\n\n" +
			tui.HelpStyle.Render("[ctrl+e] Execute  [esc] Back"),
	)
}
