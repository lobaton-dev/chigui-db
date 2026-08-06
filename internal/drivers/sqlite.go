package drivers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/lobaton-dev/chigui-db/internal/models"
	_ "modernc.org/sqlite"
)

type SQLiteDriver struct{}

func NewSQLiteDriver() *SQLiteDriver {
	return &SQLiteDriver{}
}

func (d *SQLiteDriver) Type() models.DBType {
	return models.DBSQLite
}

func (d *SQLiteDriver) Name() string {
	return "sqlite"
}

func (d *SQLiteDriver) DefaultPort() int {
	return 0
}

func (d *SQLiteDriver) BuildDSN(cfg *models.ConnectionConfig) string {
	if cfg.RawDSN != "" {
		return cfg.RawDSN
	}
	if cfg.Database == "" {
		return ":memory:"
	}
	return cfg.Database
}

func (d *SQLiteDriver) Open(cfg *models.ConnectionConfig, tlsConfig *tls.Config, dialer *net.Dialer) (any, error) {
	dsn := d.BuildDSN(cfg)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}

	for k, v := range cfg.Extra {
		pragmas = append(pragmas, fmt.Sprintf("PRAGMA %s=%s", k, v))
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			fmt.Fprintf(os.Stderr, "warn: pragma failed: %v\n", err)
		}
	}

	if cfg.Password != "" {
		if _, err := db.Exec(fmt.Sprintf("PRAGMA key='%s'", cfg.Password)); err != nil {
			return nil, fmt.Errorf("setting encryption key: %w", err)
		}
	}
	return db, nil
}

func (d *SQLiteDriver) Ping(ctx context.Context, conn any) error {
	return conn.(*sql.DB).PingContext(ctx)
}

func (d *SQLiteDriver) Close(conn any) error {
	return conn.(*sql.DB).Close()
}

func (d *SQLiteDriver) GetVersion(ctx context.Context, conn any) (string, error) {
	var v string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&v)
	return v, err
}

func (d *SQLiteDriver) GetDatabases(ctx context.Context, conn any) ([]string, error) {
	return []string{"main"}, nil
}

func (d *SQLiteDriver) GetSchemas(ctx context.Context, conn any, database string) ([]string, error) {
	return []string{"main"}, nil
}

func (d *SQLiteDriver) GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}
	var tables []*models.Table
	for rows.Next() {
		t := &models.Table{Schema: schema}
		rows.Scan(&t.Name)
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *SQLiteDriver) GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*models.Column
	for rows.Next() {
		col := &models.Column{}
		var nullable int
		var defaultVal sql.NullString
		rows.Scan(&col.Position, &col.Name, &col.Type, &nullable, &defaultVal, &col.IsPrimaryKey)
		col.Nullable = nullable == 0
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		col.Position++
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (d *SQLiteDriver) GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error) {
	return nil, nil
}

func (d *SQLiteDriver) GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error) {
	return nil, nil
}

func (d *SQLiteDriver) GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error) {
	return nil, nil
}

func (d *SQLiteDriver) GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx,
		"SELECT name, sql FROM sqlite_master WHERE type='view' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var views []*models.View
	for rows.Next() {
		v := &models.View{Schema: schema}
		rows.Scan(&v.Name, &v.Definition)
		views = append(views, v)
	}
	return views, rows.Err()
}

func (d *SQLiteDriver) GetTableDDL(ctx context.Context, conn any, schema, table string) (string, error) {
	var ddl string
	err := conn.(*sql.DB).QueryRowContext(ctx,
		"SELECT sql FROM sqlite_master WHERE name = ? AND type='table'", table).Scan(&ddl)
	return ddl, err
}

func (d *SQLiteDriver) LimitSyntax(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
func (d *SQLiteDriver) Placeholder(index int) string { return "?" }
func (d *SQLiteDriver) QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (d *SQLiteDriver) QuoteString(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
func (d *SQLiteDriver) JSONSupported() bool  { return true }
func (d *SQLiteDriver) ArraySupported() bool { return false }
