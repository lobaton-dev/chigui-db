package drivers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"strings"

	_ "github.com/godror/godror"
	"github.com/lobaton-dev/chigui-db/internal/models"
)

type OracleDriver struct{}

func NewOracleDriver() *OracleDriver {
	return &OracleDriver{}
}
func (d *OracleDriver) Type() models.DBType { return models.DBOracle }
func (d *OracleDriver) Name() string        { return "oracle" }
func (d *OracleDriver) DefaultPort() int    { return 1521 }

func (d *OracleDriver) Open(cfg *models.ConnectionConfig, tlsConfig *tls.Config, dialer *net.Dialer) (any, error) {
	dsn := d.BuildDSN(cfg)
	db, err := sql.Open("gordor", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *OracleDriver) BuildDSN(cfg *models.ConnectionConfig) string {
	if cfg.RawDSN != "" {
		return cfg.RawDSN
	}

	// Extra params
	var extraParams string
	if extra := cfg.BuildExtraDSN(); extra != "" {
		for part := range strings.SplitSeq(extra, "&") {
			if key, val, ok := strings.Cut(part, "="); ok {
				if extraParams != "" {
					extraParams += "&"
				}
				extraParams += key + "=" + val
			}
		}
	}

	// SSL: use tcps protocol
	protocol := "tcp"
	if cfg.SSLConfig != nil && cfg.SSLConfig.Mode != "" && cfg.SSLConfig.Mode != "disable" {
		protocol = "tcps"
	}

	connectStr := fmt.Sprintf(`(DESCRIPTION=(ADDRESS=(PROTOCOL=%s)(HOST=%s)(PORT=%d))(CONNECT_DATA=(SERVICE_NAME=%s)))`, protocol, cfg.Host, cfg.Port, cfg.Database)

	parts := []string{
		fmt.Sprintf(`user="%s"`, cfg.Username),
		fmt.Sprintf(`password="%s"`, cfg.Password),
		fmt.Sprintf(`connectString="%s"`, connectStr),
	}

	if extraParams != "" {
		parts = append(parts, extraParams)
	}

	return strings.Join(parts, "")
}

func (d *OracleDriver) Ping(ctx context.Context, conn any) error {
	return conn.(*sql.DB).PingContext(ctx)
}

func (d *OracleDriver) Close(conn any) error {
	return conn.(*sql.DB).Close()
}

func (d *OracleDriver) GetVersion(ctx context.Context, conn any) (string, error) {
	var v string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SELECT version FROM v$instance").Scan(&v)
	return v, err
}

func (d *OracleDriver) GetDatabases(ctx context.Context, conn any) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT name FROM v$database")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var databases []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		databases = append(databases, name)
	}
	return databases, rows.Err()
}

func (d *OracleDriver) GetSchemas(ctx context.Context, conn any, database string) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT username FROM all_users ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schemas []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		schemas = append(schemas, name)
	}
	return schemas, rows.Err()
}

func (d *OracleDriver) GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error) {
	query := `SELECT table_name, 'TABLE' FROM all_tables WHERE owner = :1 ORDER BY table_name`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []*models.Table
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &models.Table{Name: name, Schema: schema, Type: "TABLE"})
	}
	return tables, rows.Err()
}

func (d *OracleDriver) GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error) {
	query := `
        SELECT column_name, data_type, nullable, data_length
        FROM all_tab_columns
        WHERE owner = :1 AND table_name = :2
        ORDER BY column_id`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*models.Column
	for rows.Next() {
		var name, dataType string
		var nullable string
		var length int
		rows.Scan(&name, &dataType, &nullable, &length)
		col := &models.Column{Name: name, Type: dataType, Nullable: nullable == "Y"}
		if length > 0 {
			col.Type = fmt.Sprintf("%s(%d)", dataType, length)
		}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (d *OracleDriver) GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error) {
	query := `
        SELECT i.index_name, c.column_name
        FROM all_indexes i
        JOIN all_ind_columns c ON i.index_name = c.index_name AND i.table_owner = c.table_owner
        WHERE i.table_owner = :1 AND i.table_name = :2
        ORDER BY i.index_name, c.column_position`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []*models.Index
	var current *models.Index
	for rows.Next() {
		var idxName, colName string
		rows.Scan(&idxName, &colName)
		if current == nil || current.Name != idxName {
			current = &models.Index{Name: idxName, Columns: []string{}}
			indexes = append(indexes, current)
		}
		current.Columns = append(current.Columns, colName)
	}
	return indexes, rows.Err()
}

func (d *OracleDriver) GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error) {
	query := `
        SELECT
            c.constraint_name,
            col.column_name,
            r.table_name,
            r.column_name
        FROM all_cons_columns col
        JOIN all_constraints c ON col.constraint_name = c.constraint_name AND col.owner = c.owner
        JOIN all_cons_columns r ON c.r_constraint_name = r.constraint_name AND r.owner = c.r_owner
        WHERE c.constraint_type = 'R' AND col.owner = :1 AND col.table_name = :2`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []*models.ForeignKey
	for rows.Next() {
		var fk models.ForeignKey
		rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn)
		fks = append(fks, &fk)
	}
	return fks, rows.Err()
}

func (d *OracleDriver) GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error) {
	query := `SELECT trigger_name, trigger_body FROM all_triggers WHERE owner = :1 AND table_name = :2`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var triggers []*models.Trigger
	for rows.Next() {
		var t models.Trigger
		rows.Scan(&t.Name, &t.Procedure)
		triggers = append(triggers, &t)
	}
	return triggers, rows.Err()
}

func (d *OracleDriver) GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error) {
	query := `SELECT view_name, text FROM all_views WHERE owner = :1 ORDER BY view_name`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var views []*models.View
	for rows.Next() {
		var v models.View
		rows.Scan(&v.Name, &v.Definition)
		views = append(views, &v)
	}
	return views, rows.Err()
}

func (d *OracleDriver) GetTableDDL(ctx context.Context, conn any, schema, table string) (string, error) {
	query := `SELECT DBMS_METADATA.GET_DDL('TABLE', :2, :1) FROM dual`
	var ddl string
	err := conn.(*sql.DB).QueryRowContext(ctx, query, schema, table).Scan(&ddl)
	return ddl, err
}

func (d *OracleDriver) LimitSyntax(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}
	if limit > 0 {
		return fmt.Sprintf("FETCH NEXT %d ROWS ONLY", limit)
	}
	return ""
}

func (d *OracleDriver) Placeholder(index int) string {
	return fmt.Sprintf(":%d", index+1)
}

func (d *OracleDriver) QuoteIdent(name string) string {
	return "\"" + name + "\""
}

func (d *OracleDriver) QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (d *OracleDriver) JSONSupported() bool  { return true }
func (d *OracleDriver) ArraySupported() bool { return false }
