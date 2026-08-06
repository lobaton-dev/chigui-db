package drivers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/lobaton-dev/chigui-db/internal/models"
	mssql "github.com/microsoft/go-mssqldb"
)

type MSSQLDriver struct{}

func NewMSSQLDriver() *MSSQLDriver         { return &MSSQLDriver{} }
func (d *MSSQLDriver) Type() models.DBType { return models.DBMSSQL }
func (d *MSSQLDriver) Name() string        { return "sqlserver" }
func (d *MSSQLDriver) DefaultPort() int    { return 1433 }

func (d *MSSQLDriver) Open(cfg *models.ConnectionConfig, tlsConfig *tls.Config, dialer *net.Dialer) (any, error) {
	dsn := d.BuildDSN(cfg)

	var db *sql.DB
	if dialer != nil {
		connector, err := mssql.NewConnector(dsn)
		if err != nil {
			return nil, err
		}
		connector.Dialer = dialer
		db = sql.OpenDB(connector)
	} else {
		var err error
		db, err = sql.Open("sqlserver", dsn)
		if err != nil {
			return nil, err
		}
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	return db, nil
}

func (d *MSSQLDriver) BuildDSN(cfg *models.ConnectionConfig) string {
	if cfg.RawDSN != "" {
		return cfg.RawDSN
	}

	params := url.Values{}
	params.Set("database", cfg.Database)

	if cfg.Timeout > 0 {
		params.Set("connection timeout", fmt.Sprintf("%.0f", cfg.Timeout.Seconds()))
	}

	// SSL: encrypt=true/disable
	if cfg.SSLConfig != nil && cfg.SSLConfig.Mode == "disable" {
		params.Set("encrypt", "disable")
	} else {
		params.Set("encrypt", "true")
		if cfg.SSLConfig != nil && cfg.SSLConfig.ServerName != "" {
			params.Set("TrustServerCertificate", "false")
			params.Set("hostNameInCertificate", cfg.SSLConfig.ServerName)
		} else {
			params.Set("TrustServerCertificate", "true")
		}
	}

	// Extra params
	if extra := cfg.BuildExtraDSN(); extra != "" {
		for part := range strings.SplitSeq(extra, "&") {
			if key, val, ok := strings.Cut(part, "="); ok {
				params.Set(key, val)
			}
		}
	}

	// SQL Server DSN: sqlserver://user:pass@host:port?params
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?%s",
		url.QueryEscape(cfg.Username),
		url.QueryEscape(cfg.Password),
		cfg.Host, cfg.Port,
		params.Encode(),
	)
	return dsn
}

func (d *MSSQLDriver) Ping(ctx context.Context, conn any) error {
	return conn.(*sql.DB).PingContext(ctx)
}

func (d *MSSQLDriver) Close(conn any) error {
	return conn.(*sql.DB).Close()
}

func (d *MSSQLDriver) GetVersion(ctx context.Context, conn any) (string, error) {
	var v string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SELECT @@VERSION").Scan(&v)
	return v, err
}

func (d *MSSQLDriver) GetDatabases(ctx context.Context, conn any) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT name FROM sys.databases WHERE state = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

func (d *MSSQLDriver) GetSchemas(ctx context.Context, conn any, db string) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT schema_name FROM information_schema.schemata")
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

func (d *MSSQLDriver) GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error) {
	query := `SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.tables WHERE TABLE_SCHEMA = @p1`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []*models.Table
	for rows.Next() {
		var name, kind string
		rows.Scan(&name, &kind)
		tables = append(tables, &models.Table{Name: name, Schema: schema, Type: kind})
	}
	return tables, rows.Err()
}

func (d *MSSQLDriver) GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error) {
	query := `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COALESCE(CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, -1)
              FROM information_schema.columns WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2
              ORDER BY ORDINAL_POSITION`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*models.Column
	for rows.Next() {
		var name, dataType, nullable string
		var length int
		rows.Scan(&name, &dataType, &nullable, &length)
		col := &models.Column{Name: name, Type: dataType, Nullable: nullable == "YES"}
		if length > 0 {
			col.Type = fmt.Sprintf("%s(%d)", dataType, length)
		}
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (d *MSSQLDriver) GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error) {
	query := `
        SELECT i.name, i.type_desc, c.name
        FROM sys.indexes i
        JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
        JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
        WHERE i.object_id = OBJECT_ID(@p1)
        ORDER BY i.name, ic.key_ordinal`
	fullName := schema + "." + table
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, fullName)
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

func (d *MSSQLDriver) GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error) {
	query := `
        SELECT
            fk.name,
            OBJECT_NAME(fk.parent_object_id) AS source_table,
            c.name AS source_column,
            OBJECT_NAME(fk.referenced_object_id) AS target_table,
            rc.name AS target_column
        FROM sys.foreign_keys fk
        JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
        JOIN sys.columns c ON fkc.parent_object_id = c.object_id AND fkc.parent_column_id = c.column_id
        JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
        WHERE fk.parent_object_id = OBJECT_ID(@p1)`
	fullName := schema + "." + table
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, fullName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []*models.ForeignKey
	for rows.Next() {
		var fk models.ForeignKey
		var sourceTable string
		rows.Scan(&fk.Name, &sourceTable, &fk.Column, &fk.RefTable, &fk.RefColumn)
		fks = append(fks, &fk)
	}
	return fks, rows.Err()
}

func (d *MSSQLDriver) GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error) {
	query := `
        SELECT name, OBJECT_DEFINITION(object_id)
        FROM sys.triggers
        WHERE parent_id = OBJECT_ID(@p1)`
	fullName := schema + "." + table
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, fullName)
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

func (d *MSSQLDriver) GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error) {
	query := `SELECT TABLE_NAME, VIEW_DEFINITION FROM information_schema.views WHERE TABLE_SCHEMA = @p1`
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

func (d *MSSQLDriver) GetTableDDL(ctx context.Context, conn any, schema, table string) (string, error) {
	query := `SELECT OBJECT_DEFINITION(OBJECT_ID(@p1))`
	fullName := schema + "." + table
	var ddl string
	err := conn.(*sql.DB).QueryRowContext(ctx, query, fullName).Scan(&ddl)
	return ddl, err
}

func (d *MSSQLDriver) LimitSyntax(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	}
	if limit > 0 {
		return fmt.Sprintf("OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY", limit)
	}
	return ""
}

func (d *MSSQLDriver) Placeholder(index int) string {
	return fmt.Sprintf("@p%d", index+1)
}

func (d *MSSQLDriver) QuoteIdent(name string) string {
	return "[" + name + "]"
}

func (d *MSSQLDriver) QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (d *MSSQLDriver) JSONSupported() bool  { return true }
func (d *MSSQLDriver) ArraySupported() bool { return false }
