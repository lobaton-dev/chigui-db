package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/models"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Connections struct {
	config   *config.Manager
	bus      *tui.EventBus
	registry any
	list     list.Model
	width    int
	height   int
}

func NewConnections(cfg *config.Manager, bus *tui.EventBus, _ any) *Connections {
	items := make([]list.Item, 0)

	for _, connCfg := range cfg.GetConfig().Connections {
		items = append(items, connItem{config: connCfg})
	}

	items = append(items, connItem{
		title:       "New Connection...",
		description: "Add a new database connection",
		isAction:    true,
	})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	return &Connections{
		config: cfg,
		bus:    bus,
		list:   l,
	}
}

type connItem struct {
	config      *models.ConnectionConfig
	title       string
	description string
	isAction    bool
}

func (i connItem) Title() string {
	if i.isAction {
		return i.title
	}
	return fmt.Sprintf("%s (%s)", i.config.Name, i.config.Type)
}

func (i connItem) Description() string {
	if i.isAction {
		return i.description
	}
	return fmt.Sprintf("%s:%d / %s", i.config.Host, i.config.Port, i.config.Database)
}

func (i connItem) FilterValue() string {
	return i.Title() + " " + i.Description()
}

func (m *Connections) Init() tea.Cmd { return nil }

func (m *Connections) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item, ok := m.list.SelectedItem().(connItem)
			if !ok {
				return m, nil
			}
			if item.isAction {
				return m, func() tea.Msg {
					return tui.NavigateToMsg{Screen: tui.ConnectFormScreen}
				}
			}
			// Fase 1: mock — navegar directo al browser
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.BrowserScreen}
			}
			// Fase 2: reemplazar por pool.Connect + ConnSelectedMsg
		case "d":
			// Delete connection
			item, ok := m.list.SelectedItem().(connItem)
			if !ok || item.isAction {
				return m, nil
			}
			m.config.RemoveConnection(item.config.ID)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Connections) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("Connectios"))
	b.WriteString("\n\n")
	b.WriteString(m.list.View())
	b.WriteString("\n\n")
	b.WriteString(tui.HelpStyle.Render("[enter] Connect [d] Delete [esc] Back"))
	return tui.AppStyle.Render(b.String())
}
