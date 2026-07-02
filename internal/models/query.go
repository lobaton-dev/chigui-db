package models

import "time"

type Query struct {
	ID           string
	ConnectionID string
	Text         string
	Database     string
	ExecuteAt    time.Time
	Duration     time.Duration
	RowCount     int
	Error        string
	Favorite     bool
	Tags         []string
}

type QueryResult struct {
	Columns  []Column
	Rows     [][]any
	Duration time.Duration
	RowCount int
	Offset   int
	Limit    int
	Error    string
}

type QueryPlan struct {
	NodeType string
	Cost     float64
	Rows     int64
	Width    int
	Children []QueryPlan
	Actual   struct {
		Rows        int64
		Loops       int64
		StartupTime float64
		TotalTime   float64
	}
}
