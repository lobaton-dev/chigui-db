// Package tui implements the terminal UI framework: app root, styles, event bus, and all shared message types for screen navigation and data flow.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
)

type Screen int

const (
	WelcomeScreen Screen = iota
	ConnectionsScreen
	ConnectFormScreen
	BrowserScreen
	EditorScreen
	GridScreen
	ERDScreen
	AdminScreen
	HistoryScreen
	SettingsScreen
	ExportScreen
	NoSQLScreen
	MockDataScreen
	BackupScreen
	ImportScreen
	CompareScreen
	SearchScreen
	AlterTableScreen
	CreateTableScreen
)

type AppConfig struct {
	Config   *config.Manager
	Registry any
	Session  *config.Session
}

type App struct {
	config   *config.Manager
	registry any
	bus      *EventBus
	pool     any
	screens  map[Screen]tea.Model
	current  Screen
	width    int
	height   int
}

func NewApp(cfg AppConfig) *App {
	bus := NewEventBus()
	return &App{
		config:   cfg.Config,
		registry: cfg.Registry,
		bus:      bus,
		pool:     nil,
		screens:  make(map[Screen]tea.Model),
		current:  WelcomeScreen,
		width:    80,
		height:   24,
	}
}

func (a *App) RegisterScreen(s Screen, m tea.Model) {
	a.screens[s] = m
}

func (a *App) Bus() *EventBus {
	return a.bus
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NavigateToMsg:
		a.current = msg.Screen
		windowMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height}
		if s, ok := a.screens[a.current]; ok {
			var cmd tea.Cmd
			a.screens[a.current], cmd = s.Update(windowMsg)
			return a, cmd
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	if s, ok := a.screens[a.current]; ok {
		var cmd tea.Cmd
		a.screens[a.current], cmd = s.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	if s, ok := a.screens[a.current]; ok {
		return s.View()
	}
	return "chiguidb\n"
}

func (a *App) SessionState() *config.Session {
	return &config.Session{
		LastScreen:   screenName(a.current),
		WindowWidth:  a.width,
		WindowHeight: a.height,
	}
}

func screenName(s Screen) string {
	names := map[Screen]string{
		WelcomeScreen:     "welcome",
		ConnectionsScreen: "connections",
		ConnectFormScreen: "connect_form",
		BrowserScreen:     "browser",
		EditorScreen:      "editor",
		GridScreen:        "grid",
		ERDScreen:         "erd",
		AdminScreen:       "admin",
		HistoryScreen:     "history",
		SettingsScreen:    "settings",
		ExportScreen:      "export",
		NoSQLScreen:       "nosql",
		MockDataScreen:    "mockdata",
		BackupScreen:      "backup",
		ImportScreen:      "import",
		CompareScreen:     "compare",
		SearchScreen:      "search",
		AlterTableScreen:  "alter_table",
		CreateTableScreen: "create_table",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return "unknown"
}
