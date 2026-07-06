package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type ERDScreen struct {
	config  *config.Manager
	bus     *tui.EventBus
	tables  []string
	offsetX int
	offsetY int
	zoom    float64
	width   int
	height  int
}

func NewERD(cfg *config.Manager, bus *tui.EventBus) *ERDScreen {
	tables := []string{"users", "orders", "products"}
	return &ERDScreen{
		config: cfg,
		bus:    bus,
		tables: tables,
		zoom:   1.0,
	}
}

func (m *ERDScreen) Init() tea.Cmd {
	return nil
}

func (m *ERDScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.offsetX -= 5
		case "right", "l":
			m.offsetX += 5
		case "up", "k":
			m.offsetY -= 2
		case "down", "j":
			m.offsetY += 2
		case "+", "=":
			m.zoom *= 1.2
		case "-":
			m.zoom /= 1.2
		case "esc":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.BrowserScreen}
			}
		}
	}
	return m, nil
}

func (m *ERDScreen) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("ER Diagram"))
	b.WriteString("\n\n")

	for _, tbl := range m.tables {
		fmt.Fprintf(&b, "┌%s┐\n", strings.Repeat("─", len(tbl)+4))
		fmt.Fprintf(&b, "│  %s  │\n", tbl)
		fmt.Fprintf(&b, "└%s┘\n\n", strings.Repeat("─", len(tbl)+4))
	}

	b.WriteString(tui.HelpStyle.Render("[arrows] Pan  [+/-] Zoom  [esc] Back"))
	return tui.AppStyle.Render(b.String())
}
