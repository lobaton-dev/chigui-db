package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Grid struct {
	config    *config.Manager
	bus       *tui.EventBus
	columns   []string
	rows      [][]string
	cursorRow int
	cursorCol int
	width     int
	height    int
}

func NewGrid(cfg *config.Manager, bus *tui.EventBus) *Grid {
	columns := []string{"id", "name", "email", "role", "created_at"}
	rows := [][]string{
		{"1", "Alice", "alice@example.com", "admin", "2024-01-15"},
		{"2", "Bob", "bob@example.com", "editor", "2024-02-20"},
		{"3", "Carol", "carol@example.com", "viewer", "2024-03-10"},
		{"4", "Dave", "dave@example.com", "manager", "2024-04-05"},
		{"5", "Eve", "eve@example.com", "editor", "2024-05-01"},
	}
	return &Grid{config: cfg, bus: bus, columns: columns, rows: rows}
}

func (m *Grid) Init() tea.Cmd {
	return nil
}

func (m *Grid) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursorRow > 0 {
				m.cursorRow--
			}
		case "down", "j":
			if m.cursorRow < len(m.rows)-1 {
				m.cursorRow++
			}
		case "left", "h":
			if m.cursorCol > 0 {
				m.cursorCol--
			}
		case "right", "l":
			if m.cursorCol < len(m.columns)-1 {
				m.cursorCol++
			}
		case "esc":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.BrowserScreen}
			}
		}
	}
	return m, nil
}

func (m *Grid) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("Data Browser"))
	b.WriteString("\n\n")

	for _, col := range m.columns {
		fmt.Fprintf(&b, "%-15s", col)
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n")

	for i, row := range m.rows {
		prefix := " "
		if i == m.cursorRow {
			prefix = "> "
		}
		b.WriteString(prefix)
		for ci, cell := range row {
			val := fmt.Sprintf("%-15s", cell)
			if i == m.cursorRow && ci == m.cursorCol {
				b.WriteString(tui.SelectedStyle.Render(val))
			} else {
				fmt.Fprintf(&b, "%-15s", cell)
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(tui.HelpStyle.Render("[↑/↓/←/→] Navigate  [esc] Back"))
	return tui.AppStyle.Render(b.String())
}
