package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type History struct {
	config *config.Manager
	bus    *tui.EventBus
	items  []string
	cursor int
	width  int
	height int
}

func NewHistory(cfg *config.Manager, bus *tui.EventBus) *History {
	items := []string{
		"SELECT * FROM users",
		"INSERT INTO orders (user_id, total) VALUES (1, 99.99)",
		"UPDATE products SET stock = stock - 1 WHERE id = 42",
		"DELETE FROM sessions WHERE expires_at < NOW()",
	}
	return &History{config: cfg, bus: bus, items: items}
}

func (m *History) Init() tea.Cmd {
	return nil
}

func (m *History) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "esc":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.BrowserScreen}
			}
		}
	}
	return m, nil
}

func (m *History) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("Query History"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		prefix := " "
		if i == m.cursor {
			prefix = "> "
		}
		fmt.Fprintf(&b, "%s%s\n", prefix, item)
	}
	b.WriteString("\n")
	b.WriteString(tui.HelpStyle.Render("[↑/↓] Navigate  [esc] Back"))
	return tui.AppStyle.Render(b.String())
}
