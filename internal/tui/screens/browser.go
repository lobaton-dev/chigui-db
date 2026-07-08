package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
)

type Browser struct {
	config  *config.Manager
	bus     *tui.EventBus
	spinner spinner.Model
	tree    []treeNode
	cursor  int
	loading bool
	width   int
	height  int
}

type treeNode struct {
	name     string
	kind     string
	children []treeNode
	expanded bool
	depth    int
}

func NewBrowser(cfg *config.Manager, bus *tui.EventBus) *Browser {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = tui.SelectedStyle

	return &Browser{
		config:  cfg,
		bus:     bus,
		spinner: s,
		loading: true,
	}
}

func (m *Browser) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadSchema())
}

func (m *Browser) loadSchema() tea.Cmd {
	return func() tea.Msg {
		return tui.SchemaLoadedMsg{}
	}
}

func (m *Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tree)-1 {
				m.cursor++
			}
		case "right", "l":
			if m.cursor < len(m.tree) && !m.tree[m.cursor].expanded {
				m.tree[m.cursor].expanded = true
				m.rebuildFlat()
			}
		case "left", "h":
			if m.cursor < len(m.tree) && m.tree[m.cursor].expanded {
				m.tree[m.cursor].expanded = false
				m.rebuildFlat()
			}
		case "enter":
			if m.cursor < len(m.tree) && m.tree[m.cursor].kind == "table" {
				return m, func() tea.Msg {
					return tui.NavigateToMsg{
						Screen: tui.GridScreen,
						Data:   m.tree[m.cursor].name,
					}
				}
			}
		case "e":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.EditorScreen}
			}
		case "r":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.ERDScreen}
			}
		case "esc":
			return m, func() tea.Msg {
				return tui.NavigateToMsg{Screen: tui.WelcomeScreen}
			}
		}
	}
	return m, nil
}

func (m *Browser) rebuildFlat() {
	var flat []treeNode
	var walk func(nodes []treeNode, depth int)
	walk = func(nodes []treeNode, depth int) {
		for _, node := range nodes {
			node.depth = depth
			flat = append(flat, node)
			if node.expanded && len(node.children) > 0 {
				walk(node.children, depth+1)
			}
		}
	}
	walk(m.tree, 0)
	m.tree = flat
	if m.cursor >= len(m.tree) {
		m.cursor = len(m.tree) - 1
	}
}

func (m *Browser) View() string {
	if m.loading {
		return tui.AppStyle.Render(
			tui.TitleStyle.Render("Schema Browser") + "\n\n" +
				m.spinner.View() + " Loading schema...",
		)
	}
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Render("Schema Browser"))
	b.WriteString("\n\n")

	for i, node := range m.tree {
		prefix := ""
		if i == m.cursor {
			prefix = "> "
		} else {
			prefix = "  "
		}
		indent := strings.Repeat("  ", node.depth)
		icon := iconForKind(node.kind)
		expand := " "
		if len(node.children) > 0 {
			if node.expanded {
				expand = "▼"
			} else {
				expand = "▶"
			}
		}
		b.WriteString(prefix)
		b.WriteString(indent)
		b.WriteString(expand)
		b.WriteString(" ")
		b.WriteString(icon)
		b.WriteString(" ")
		b.WriteString(node.name)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(tui.HelpStyle.Render("[↑/↓] Navigate  [→] Expand  [←] Collapse  [enter] View data  [e] SQL  [r] ERD"))
	return tui.AppStyle.Render(b.String())
}

func iconForKind(kind string) string {
	switch kind {
	case "database":
		return "🗄"
	case "schema":
		return "📁"
	case "table":
		return "📋"
	case "view":
		return "👁"
	case "column":
		return "▸"
	case "index":
		return "🔍"
	case "trigger":
		return "⚡"
	default:
		return "•"
	}
}
