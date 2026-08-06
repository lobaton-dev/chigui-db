// Package drivers provides database driver interfaces and registry for both SQL databases
package drivers

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/lobaton-dev/chigui-db/internal/models"
)

type Driver interface {
	Type() models.DBType
	Name() string
	DefaultPort() int
	BuildDSN(config *models.ConnectionConfig) string
	Open(cfg *models.ConnectionConfig, tls *tls.Config, dialer *net.Dialer) (any, error)
	Ping(ctx context.Context, conn any) error
	Close(conn any) error
	GetVersion(ctx context.Context, conn any) (string, error)
	GetDatabases(ctx context.Context, conn any) ([]string, error)
	GetSchemas(ctx context.Context, conn any, database string) ([]string, error)
	GetTables(ctx context.Context, conn any, schema string) ([]*models.Table, error)
	GetColumns(ctx context.Context, conn any, schema, table string) ([]*models.Column, error)
	GetIndexes(ctx context.Context, conn any, schema, table string) ([]*models.Index, error)
	GetForeignKeys(ctx context.Context, conn any, schema, table string) ([]*models.ForeignKey, error)
	GetTriggers(ctx context.Context, conn any, schema, table string) ([]*models.Trigger, error)
	GetViews(ctx context.Context, conn any, schema string) ([]*models.View, error)
	GetTableDDL(ctx context.Context, con any, schema, table string) (string, error)
	LimitSyntax(limit, offset int) string
	Placeholder(index int) string
	QuoteIdent(name string) string
	QuoteString(s string) string
	JSONSupported() bool
	ArraySupported() bool
}
