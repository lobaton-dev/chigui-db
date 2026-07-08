package screens

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/models"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type ConnectForm struct {
	config   *config.Manager
	bus      *tui.EventBus
	registry any
	form     *huh.Form
	editing  bool
	editID   string

	name     string
	dbType   string
	host     string
	port     string
	database string
	user     string
	password string
}

func NewConnectForm(cfg *config.Manager, bus *tui.EventBus, _ any) *ConnectForm {
	s := &ConnectForm{
		config: cfg,
		bus:    bus,
	}

	typeOpts := []huh.Option[string]{
		huh.NewOption("PostgreSQL", "postgresql"),
		huh.NewOption("MySQL", "mysql"),
		huh.NewOption("SQLite", "sqlite"),
		huh.NewOption("SQL Server", "mssql"),
		huh.NewOption("Oracle", "oracle"),
		huh.NewOption("ClickHouse", "clickhouse"),
		huh.NewOption("CockroachDB", "cockroach"),
		huh.NewOption("MongoDB", "mongodb"),
		huh.NewOption("Redis", "redis"),
		huh.NewOption("Cassandra", "cassandra"),
		huh.NewOption("Neo4j", "neo4j"),
		huh.NewOption("DynamoDB", "dynamodb"),
	}

	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Connection Name").Key("name").Value(&s.name),
			huh.NewSelect[string]().Title("Database Type").Key("type").Options(typeOpts...).Value(&s.dbType),
			huh.NewInput().Title("Host").Key("host").Value(&s.host),
			huh.NewInput().Title("Port").Key("port").Value(&s.port),
			huh.NewInput().Title("Database").Key("database").Value(&s.database),
			huh.NewInput().Title("Username").Key("user").Value(&s.user),
			huh.NewInput().Title("Password").Key("password").EchoMode(huh.EchoModePassword).Value(&s.password),
		),
	)
	return s
}

func (m *ConnectForm) Init() tea.Cmd { return m.form.Init() }

func (m *ConnectForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.ConnectionsScreen}
			}
		}
	}

	form, cmd := m.form.Update(msg)
	m.form = form.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		cfg, err := m.buildConfig()
		if err != nil {
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.ConnectionsScreen}
			}
		}
		m.config.AddConnection(cfg)
		_ = m.config.Save()

		return m, func() tea.Msg {
			return tui.NavigateToMsg{Screen: tui.ConnectionsScreen}
		}
	}
	return m, cmd
}

func (m *ConnectForm) buildConfig() (*models.ConnectionConfig, error) {
	id := fmt.Sprintf("%x", md5.Sum(fmt.Appendf(nil, "%s:%s:%s", m.dbType, m.host, m.database)))
	port, err := strconv.Atoi(m.port)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", m.port, err)
	}
	return &models.ConnectionConfig{
		ID:        id,
		Name:      m.name,
		Type:      models.DBType(m.dbType),
		Host:      m.host,
		Port:      port,
		Database:  m.database,
		Username:  m.user,
		Password:  m.password,
		Timeout:   10 * time.Second,
		MaxOpen:   10,
		MaxIdle:   5,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *ConnectForm) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render(m.editTitle()))
	b.WriteString("\n\n")
	b.WriteString(m.form.View())
	b.WriteString("\n\n")
	b.WriteString(tui.HelpStyle.Render("[enter] Save  [esc] Cancel"))
	return tui.AppStyle.Render(b.String())
}

func (m *ConnectForm) editTitle() string {
	if m.editing {
		return "Edit Connection"
	}
	return "New Connection"
}
