package models

type Database struct {
	Name    string
	Owner   string
	Size    int64
	Schemas []Schema
}

type Schema struct {
	Name   string
	Owner  string
	Tables []Table
	Views  []View
}

type Table struct {
	Name        string
	Schema      string
	Type        string
	Owner       string
	RowCount    int64
	Size        int64
	Columns     []Column
	Indexes     []Index
	Triggers    []Trigger
	ForeignKeys []ForeignKey
	DDL         string
}

type Column struct {
	Name          string
	Position      int
	Type          string
	Nullable      bool
	Default       string
	IsPrimaryKey  bool
	IsForeignKey  bool
	IsUnique      bool
	AutoIncrement bool
	Comment       string
	CharMaxLength int64
	NumericPrec   int
	NumericScale  int
}

type Index struct {
	Name       string
	Unique     bool
	Primary    bool
	Type       string
	Columns    []string
	Definition string
}

type Trigger struct {
	Name      string
	Table     string
	Event     string
	Timing    string
	Level     string
	Procedure string
}

type ForeignKey struct {
	Name      string
	Column    string
	RefTable  string
	RefColumn string
	OnDelete  string
	OnUpdate  string
}

type View struct {
	Name       string
	Schema     string
	Definition string
	Columns    []Column
}
