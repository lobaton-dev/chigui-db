// Package screens contains all application screen implementations.
package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Welcome struct {
	config *config.Manager
	bus    *tui.EventBus
	list   list.Model
	width  int
	height int
}

type menuItem struct {
	title       string
	description string
	screen      tui.Screen
}

func (i menuItem) Title() string {
	return i.title
}

func (i menuItem) Description() string {
	return i.description
}

func (i menuItem) FilterValue() string {
	return i.title
}

func NewWelcome(cfg *config.Manager, bus *tui.EventBus) *Welcome {
	items := []list.Item{
		menuItem{"Connect to Database", "Open a new database", tui.ConnectionsScreen},
		menuItem{"Query History", "Browse recent executed queries", tui.HistoryScreen},
		menuItem{"Settings", "Configure chiguidb", tui.SettingsScreen},
		menuItem{"Quit", "Exit chiguidb", tui.WelcomeScreen},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "ChiguiDB"
	l.SetShowHelp(true)

	return &Welcome{config: cfg, bus: bus, list: l}
}

func (m *Welcome) Init() tea.Cmd {
	return nil
}

func (m *Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item, ok := m.list.SelectedItem().(menuItem)
			if !ok {
				return m, nil
			}
			if item.title == "Quit" {
				return m, tea.Quit
			}
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: item.screen}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Welcome) View() string {
	return tui.AppStyle.Render(
		fmt.Sprintf("%s\n\n%s",
			tui.TitleStyle.Render(" CHIGUIDB "),
			m.list.View(),
		),
	)
}
