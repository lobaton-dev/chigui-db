// Package main is the application entry point.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lobaton-dev/chigui-db/internal/config"
	"github.com/lobaton-dev/chigui-db/internal/tui"
	"github.com/lobaton-dev/chigui-db/internal/tui/screens"
)

var version = "0.1.0"

func main() {
	showHelp := flag.Bool("help", false, "show help")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showHelp {
		fmt.Println(` chiguidb - Database Manager TUI

Usage:
    chiguidb [flags]

Flags:
    --help       Show this help
    --version    Show version

Controls:
    ↑/k ↓/j     Navigate
    Enter       Select
    Esc         Back
    Ctrl+C/q    Quit`)
		return
	}

	if *showVersion {
		fmt.Println("chiguidb", version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	session, _ := cfg.LoadSession()

	app := tui.NewApp(tui.AppConfig{
		Config:   cfg,
		Registry: nil, // *drivers.Registry — Fase 2
		Session:  session,
	})

	bus := app.Bus()
	app.RegisterScreen(tui.WelcomeScreen, screens.NewWelcome(cfg, bus))
	app.RegisterScreen(tui.ConnectionsScreen, screens.NewConnections(cfg, bus, nil))
	app.RegisterScreen(tui.ConnectFormScreen, screens.NewConnectForm(cfg, bus, nil))
	app.RegisterScreen(tui.BrowserScreen, screens.NewBrowser(cfg, bus))
	app.RegisterScreen(tui.EditorScreen, screens.NewEditor(cfg, bus))
	app.RegisterScreen(tui.GridScreen, screens.NewGrid(cfg, bus))
	app.RegisterScreen(tui.HistoryScreen, screens.NewHistory(cfg, bus))
	app.RegisterScreen(tui.ExportScreen, screens.NewExport(cfg, bus))
	app.RegisterScreen(tui.ERDScreen, screens.NewERD(cfg, bus))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	p := tea.NewProgram(app, tea.WithAltScreen())

	go func() {
		<-sigCh
		_ = cfg.SaveSession(app.SessionState())
		os.Exit(0)
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	_ = cfg.SaveSession(app.SessionState())
}
