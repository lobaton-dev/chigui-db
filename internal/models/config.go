// Package models defines core data structures shared across all packages: connections, schemas, queries, configuration and type mappings.
package models

type Config struct {
	Connections []*ConnectionConfig `yaml:"connections"`
	UI          UIConfig            `yaml:"ui"`
	Editor      EditorConfig        `yaml:"editor"`
	Export      ExportConfig        `yaml:"export"`
}

type UIConfig struct {
	Theme     string `yaml:"theme"`
	LimitRows int    `yaml:"limit_rows"`
	FontSize  int    `yaml:"font_size"`
}

type EditorConfig struct {
	FontSize     int    `yaml:"font_size"`
	TabSize      int    `yaml:"tab_size"`
	WordWrap     bool   `yaml:"word_wrap"`
	AutoComplete bool   `yaml:"auto_complete"`
	Theme        string `yaml:"theme"`
}

type ExportConfig struct {
	DefaultFormat string `yaml:"default_format"`
	DefaultDir    string `yaml:"default_dir"`
	IncludeHeader bool   `yaml:"include_header"`
	Delimiter     string `yaml:"delimiter"`
}

func DefaultConfig() *Config {
	return &Config{
		UI: UIConfig{
			Theme:     "auto",
			LimitRows: 1000,
		},
		Editor: EditorConfig{
			TabSize:      4,
			WordWrap:     true,
			AutoComplete: true,
			Theme:        "monokai",
		},
		Export: ExportConfig{
			DefaultFormat: "csv",
			DefaultDir:    "~/exports",
			IncludeHeader: true,
			Delimiter:     ",",
		},
	}
}
