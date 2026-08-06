package drivers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lobaton-dev/chigui-db/internal/models"
)

type MySQLDriver struct{}

func NewMySQLDriver() *MySQLDriver {
	return &MySQLDriver{}
}

func (d *MySQLDriver) Type() models.DBType {
	return models.DBMySQL
}

func (d *MySQLDriver) Name() string {
	return "mysql"
}

func (d *MySQLDriver) DefaultPort() int {
	return 3306
}

func (d *MySQLDriver) BuildDSN(cfg *models.ConnectionConfig) string {
	if cfg.RawDSN != "" {
		return cfg.RawDSN
	}

	network := "tcp"
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	if cfg.SocketPath != "" {
		network = "unix"
		address = cfg.SocketPath
	}

	params := url.Values{}
	params.Set("charset", "utf8mb4")
	params.Set("parseTime", "true")
	params.Set("loc", "Local")

	if cfg.Timeout > 0 {
		params.Set("timeout", fmt.Sprintf("%.0fs", cfg.Timeout.Seconds()))
	}

	if cfg.SSLConfig != nil && cfg.SSLConfig.Mode == "true" {
		params.Set("tls", "true")
	} else if cfg.SSLConfig != nil && cfg.SSLConfig.Mode == "skip-verify" {
		params.Set("tls", "skip-verify")
	} else if cfg.SSLConfig != nil && cfg.SSLConfig.CAFile != "" {
		params.Set("tls", "custom")
	}

	if extra := cfg.BuildExtraDSN(); extra != "" {
		for part := range strings.SplitSeq(extra, "&") {
			if key, val, ok := strings.Cut(part, "="); ok {
				params.Set(key, val)
			}
		}
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s?%s",
		cfg.Username,
		url.QueryEscape(cfg.Password),
		network, address,
		cfg.Database,
		params.Encode(),
	)
	return dsn
}

func (d *MySQLDriver) Open(cfg *models.ConnectionConfig, tlsConfig *tls.Config, dialer *net.Dialer) (any, error) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = cfg.Username
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port))
	mysqlCfg.DBName = cfg.Database
	mysqlCfg.Params = map[string]string{"charset": "utf8mb4"}
	mysqlCfg.Loc = time.Local

	if cfg.SocketPath != "" {
		mysqlCfg.Net = "unix"
		mysqlCfg.Addr = cfg.SocketPath
	}

	if cfg.Timeout > 0 {
		mysqlCfg.Timeout = cfg.Timeout
	}

	if tlsConfig != nil {
		tlsKey := "chiguidb-custom"
		mysql.RegisterTLSConfig(tlsKey, tlsConfig)
		mysqlCfg.TLSConfig = tlsKey
	} else if cfg.SSLConfig != nil && cfg.SSLConfig.Mode == "skip-verify" {
		mysqlCfg.TLSConfig = "skip-verify"
	} else if cfg.SSLConfig != nil && cfg.SSLConfig.Mode == "true" {
		mysqlCfg.TLSConfig = "true"
	}

	if dialer != nil {
		mysqlCfg.DialFunc = dialer.DialContext
	}

	connector, err := mysql.NewConnector(mysqlCfg)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	return db, nil
}

func (d *MySQLDriver) Ping(ctx context.Context, conn any) error {
	return conn.(*sql.DB).PingContext(ctx)
}

func (d *MySQLDriver) Close(conn any) error {
	return conn.(*sql.DB).Close()
}

func (d *MySQLDriver) GetVersion(ctx context.Context, conn any) (string, error) {
	var v string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SELECT VERSION").Scan(&v)
	return v, err
}

func (d *MySQLDriver) GetDatabases(ctx context.Context, conn any) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SHOW DATABSES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if name != "information_schema" && name != "perfomance_schema" && name != "sys" && name != "mysql" {
			dbs = append(dbs, name)
		}
	}
	return dbs, rows.Err()
}

func (d *MySQLDriver) GetSchemas(ctx context.Context, conn any, databse string) ([]string, error) {
	return []string{databse}, nil
}

func (d *MySQLDriver) GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME", schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []*models.Table
	for rows.Next() {
		t := &models.Table{Schema: schema}
		var tableType string
		rows.Scan(&t.Name, &tableType)
		if tableType == "VIEW" {
			t.Type = "VIEW"
		} else {
			t.Type = "TABLE"
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *MySQLDriver) GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, `
        SELECT COLUMN_NAME, ORDINAL_POSITION, DATA_TYPE,
               IF(IS_NULLABLE = 'YES', TRUE, FALSE),
               COLUMN_DEFAULT, CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE,
               IF(COLUMN_KEY = 'PRI', TRUE, FALSE),
               IF(EXTRA LIKE '%auto_increment%', TRUE, FALSE)
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
        ORDER BY ORDINAL_POSITION`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*models.Column
	for rows.Next() {
		col := &models.Column{}
		rows.Scan(&col.Name, &col.Position, &col.Type, &col.Nullable,
			&col.Default, &col.CharMaxLength, &col.NumericPrec,
			&col.NumericScale, &col.IsPrimaryKey, col.AutoIncrement)
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (d *MySQLDriver) GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, `
		SELECT INDEX_NAME, COLUMN_NAME, NON_UNIQUE, SEQ_IN_INDEX, INDEX_TYPE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ODER BY INDEX_NAME, SEQ_IN_INDEX`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []*models.Index
	indexMap := make(map[string]*models.Index)
	for rows.Next() {
		var name, colName, indexType string
		var nonUnique, seqInIndex int
		rows.Scan(&name, &colName, &nonUnique, seqInIndex, indexType)
		idx, ok := indexMap[name]
		if !ok {
			idx = &models.Index{Name: name, Unique: nonUnique == 0, Type: indexType}
			indexMap[name] = idx
			indexes = append(indexes, idx)
		}
		idx.Columns = append(idx.Columns, colName)
	}
	return indexes, rows.Err()
}

func (d *MySQLDriver) GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, `
		SELECT k.CONSTRAINT_NAME, k.COLUMN_NAME, k.REFERENCED_TABLE_NAME, k.REFERENCED_COLUMN_NAME,
			r.UPDATE_RULE, r.DELETE_RULE
		FROM information_schema.KEY_COLUMN_USAGE k
		JOIN information_schema.REFERENTIAL_CONSTRAINTS r
			ON k.CONSTRAINT_NAME = r.CONSTRAINT_NAME AND k.CONSTRAINT_SCHEMA = r.CONSTRAINT_SCHEMA
		WHERE k.TABLE_SCHEMA = ? AND k.REFERENCED_TABLE_NAME IS NOT NULL`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []*models.ForeignKey
	for rows.Next() {
		fk := &models.ForeignKey{}
		rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn, &fk.OnDelete)
	}
	return fks, rows.Err()
}

func (d *MySQLDriver) GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, `
		SELECT TRIGGER_NAME, EVENT_MANIPULATION, ACTION_TIMING
		FROM information_schema.TRIGGERS
		WHERE EVENT_OBJECT_SCHEMA = ? AND EVENT_OBJECT_TABLE`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var triggers []*models.Trigger
	for rows.Next() {
		t := &models.Trigger{Table: table}
		rows.Scan(&t.Name, &t.Timing)
		triggers = append(triggers, t)
	}
	return triggers, rows.Err()
}

func (d *MySQLDriver) GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, `
		SELECT TABLE_NAME, VIEW_DEFINITION FROM information_schema.VIEWS WHERE TABLE_SCHEMA = ?`, schema)
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

func (d *MySQLDriver) GetTableDDL(ctx context.Context, conn any, schema, table string) (string, error) {
	var ddl string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SHOW CREATE TABLE `"+schema+"`.`"+table+"`").Scan(&table, &ddl)
	return ddl, err
}

func (d *MySQLDriver) LimitSyntax(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

func (d *MySQLDriver) Placeholder(index int) string {
	return "?"
}

func (d *MySQLDriver) QuoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (d *MySQLDriver) QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (d *MySQLDriver) JSONSupported() bool {
	return true
}

func (d *MySQLDriver) ArraySupported() bool {
	return false
}
