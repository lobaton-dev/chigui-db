package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Export struct {
	config *config.Manager
	bus    *tui.EventBus
	format int
	path   string
	width  int
	height int
}

var exportFormats = []string{"CSV", "JSON", "SQL", "Excel"}

func NewExport(cfg *config.Manager, bus *tui.EventBus) *Export {
	return &Export{
		config: cfg,
		bus:    bus,
		format: 0,
	}
}

func (m *Export) Init() tea.Cmd {
	return nil
}

func (m *Export) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.format > 0 {
				m.format--
			}
		case "down", "j":
			if m.format < len(exportFormats)-1 {
				m.format++
			}
		case "enter":
			return m, nil
		case "esc":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.BrowserScreen}
			}
		}
	}
	return m, nil
}

func (m *Export) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("Export Data"))
	b.WriteString("\n\n")

	for i, f := range exportFormats {
		if i == m.format {
			fmt.Fprintf(&b, "  ◉ %s\n", f)
		} else {
			fmt.Fprintf(&b, "  ○ %s\n", f)
		}
	}

	b.WriteString("\n")
	b.WriteString(tui.HelpStyle.Render("[↑/↓] Select format  [enter] Export  [esc] Back"))
	return tui.AppStyle.Render(b.String())
}
