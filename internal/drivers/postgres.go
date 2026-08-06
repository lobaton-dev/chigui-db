package drivers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/lobaton-dev/chigui-db/internal/models"
)

type PostgresDriver struct{}

type pqDialerAdapter struct {
	d *net.Dialer
}

func (a pqDialerAdapter) Dial(network, address string) (net.Conn, error) {
	return a.d.Dial(network, address)
}

func (a pqDialerAdapter) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return a.d.DialContext(ctx, network, address)
}

func NewPostgresDriver() *PostgresDriver {
	return &PostgresDriver{}
}

func (d *PostgresDriver) Type() models.DBType { return models.DBPostgreSQL }

func (d *PostgresDriver) Name() string { return "postgres" }

func (d *PostgresDriver) DefaultPort() int { return 5432 }

func (d *PostgresDriver) Open(cfg *models.ConnectionConfig, tlsConfig *tls.Config, dialer *net.Dialer) (any, error) {
	dsn := d.BuildDSN(cfg)
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	if tlsConfig != nil {
		tlsMode := cfg.SSLConfig.Mode
		if tlsMode == "" {
			tlsMode = "require"
		}
		pq.RegisterTLSConfig(tlsMode, tlsConfig)
	}

	var db *sql.DB
	if dialer != nil {
		connector, err := pq.NewConnector(dsn)
		if err != nil {
			return nil, err
		}
		connector.Dialer(pqDialerAdapter{d: dialer})
		db = sql.OpenDB(connector)
	} else {
		var err error
		db, err = sql.Open("postgres", dsn)
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

func (d *PostgresDriver) BuildDSN(cfg *models.ConnectionConfig) string {
	if cfg.RawDSN != "" {
		return cfg.RawDSN
	}

	params := url.Values{}

	sslMode := "disable"
	if cfg.SSLConfig != nil && cfg.SSLConfig.Mode != "" {
		sslMode = cfg.SSLConfig.Mode
	}

	params.Set("sslmode", sslMode)

	if cfg.SSLConfig != nil {
		if cfg.SSLConfig.CAFile != "" {
			params.Set("sslrootcert", cfg.SSLConfig.CAFile)
		}
		if cfg.SSLConfig.CertFile != "" {
			params.Set("sslcert", cfg.SSLConfig.CertFile)
		}
		if cfg.SSLConfig.KeyFile != "" {
			params.Set("sslkey", cfg.SSLConfig.KeyFile)
		}
	}

	if cfg.Timeout > 0 {
		params.Set("connect_timeout", fmt.Sprintf("%.0f", cfg.Timeout.Seconds()))
	}

	if extra := cfg.BuildExtraDSN(); extra != "" {
		for part := range strings.SplitSeq(extra, "&") {
			if key, val, ok := strings.Cut(part, "="); ok {
				params.Set(key, val)
			}
		}
	}

	host := cfg.Host
	if cfg.SocketPath != "" {
		host = cfg.SocketPath
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s %s", host,
			url.QueryEscape(cfg.Username),
			url.QueryEscape(cfg.Password),
			url.QueryEscape(cfg.Database),
			params.Encode(),
		)
		return dsn
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s", host, cfg.Port,
		url.QueryEscape(cfg.Username),
		url.QueryEscape(cfg.Password),
		url.QueryEscape(cfg.Database),
		params.Encode(),
	)
	return dsn
}

func (d *PostgresDriver) Ping(ctx context.Context, conn any) error {
	return conn.(*sql.DB).PingContext(ctx)
}

func (d *PostgresDriver) Close(conn any) error {
	return conn.(*sql.DB).Close()
}

func (d *PostgresDriver) GetVersion(ctx context.Context, conn any) (string, error) {
	var v string
	err := conn.(*sql.DB).QueryRowContext(ctx, "SELECT version()").Scan(&v)
	return v, err
}

func (d *PostgresDriver) GetDatabases(ctx context.Context, conn any) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
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

func (d *PostgresDriver) GetSchemas(ctx context.Context, conn any, database string) ([]string, error) {
	rows, err := conn.(*sql.DB).QueryContext(ctx, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
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

func (d *PostgresDriver) GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error) {
	query := `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = $1 ORDER BY tablename`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []*models.Table
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &models.Table{
			Name:   name,
			Schema: schema,
			Type:   "TABLE",
		})
	}
	return tables, rows.Err()
}

func (d *PostgresDriver) GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error) {
	query := `
        SELECT
            c.column_name,
            c.ordinal_position,
            c.data_type,
            c.is_nullable = 'YES',
            c.column_default,
            c.character_maximum_length,
            c.numeric_precision,
            c.numeric_scale,
            COALESCE(pk.is_pk, false),
            c.is_identity = 'YES'
        FROM information_schema.columns c
        LEFT JOIN (
            SELECT ku.column_name, true AS is_pk
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
            WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = $1 AND tc.table_name = $2
        ) pk ON pk.column_name = c.column_name
        WHERE c.table_schema = $1 AND c.table_name = $2
        ORDER BY c.ordinal_position`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []*models.Column
	for rows.Next() {
		col := &models.Column{}
		rows.Scan(&col.Name, &col.Position, &col.Type, &col.Nullable, &col.Default, &col.CharMaxLength, &col.NumericPrec, &col.NumericScale, &col.IsPrimaryKey, &col.AutoIncrement)
		cols = append(cols, col)
	}
	return cols, rows.Err()
}

func (d *PostgresDriver) GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error) {
	query := ` 
		SELECT i.incexname, i.indexdef
		FROM pg_indexes i
		WHERE i.schemaname = $1 AND i.tablename = $2
		ORDER BY i.indexname`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []*models.Index
	for rows.Next() {
		idx := &models.Index{Columns: []string{}}
		rows.Scan(&idx.Name, &idx.Definition)
		if strings.Contains(idx.Definition, "UNIQUE") {
			idx.Unique = true
		}
		if strings.Contains(idx.Definition, "PRIMARY KEY") {
			idx.Primary = true
		}
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (d *PostgresDriver) GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error) {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS ref_table,
			ccu.column_name AS ref_column,
			rc.update_rule,
			rc.delete_rule
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
    JOIN information_schema.referential_constraints rc ON tc.constraint_name = rc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = $1 AND tc.table_name = $2`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []*models.ForeignKey
	for rows.Next() {
		fk := &models.ForeignKey{}
		rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn, &fk.OnUpdate, &fk.OnDelete)
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (d *PostgresDriver) GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error) {
	query := `
        SELECT trigger_name, event_manipulation, action_timing, action_orientation
        FROM information_schema.triggers
        WHERE event_object_schema = $1 AND event_object_table = $2
        ORDER BY trigger_name`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var triggers []*models.Trigger
	for rows.Next() {
		t := &models.Trigger{Table: table}
		rows.Scan(&t.Name, &t.Event, &t.Timing, &t.Level)
		triggers = append(triggers, t)
	}
	return triggers, rows.Err()
}

func (d *PostgresDriver) GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error) {
	query := `SELECT table_name, view_definition FROM information_schema.views WHERE table_schema = $1 ORDER BY table_name`
	rows, err := conn.(*sql.DB).QueryContext(ctx, query, schema)
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

func (d *PostgresDriver) GetTableDDL(ctx context.Context, conn any, schema, table string) (string, error) {
	var ddl string
	err := conn.(*sql.DB).QueryRowContext(ctx,
		`SELECT pg_catalog.pg_get_viewdef(c.oid, true)
         FROM pg_catalog.pg_class c
         JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
         WHERE n.nspname = $1 AND c.relname = $2`, schema, table).Scan(&ddl)
	if err != nil {
		err = conn.(*sql.DB).QueryRowContext(ctx,
			`SELECT
						'CREATE TABLE ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname) || ' (' ||
						array_to_string(array_agg(
								quote_ident(a.attname) || ' ' ||
								pg_catalog.format_type(a.atttypid, a.atttypmod) ||
								case when a.attnotnull then ' NOT NULL' else '' end ||
								case when ad.adsrc is not null then ' DEFAULT ' || ad.adsrc else '' end
								ORDER BY a.attnum
						), ', ') || ');'
					FROM pg_catalog.pg_class c
					JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
					JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid
					LEFT JOIN pg_catalog.pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
					WHERE n.nspname = $1 AND c.relname = $2
						AND a.attnum > 0 AND NOT a.attisdropped
					GROUP BY n.nspname, c.relname`, schema, table).Scan(&ddl)
	}
	return ddl, err
}

func (d *PostgresDriver) LimitSyntax(limit, offset int) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
func (d *PostgresDriver) Placeholder(index int) string { return fmt.Sprintf("$%d", index+1) }
func (d *PostgresDriver) QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (d *PostgresDriver) QuoteString(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
func (d *PostgresDriver) JSONSupported() bool  { return true }
func (d *PostgresDriver) ArraySupported() bool { return true }
