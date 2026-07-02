package models

import "time"

type DBType string

const (
	DBPostgreSQL DBType = "postgresql"
	DBMySQL      DBType = "mysql"
	DBSQLite     DBType = "sqlite"
	DBMSSQL      DBType = "mssql"
	DBOracle     DBType = "oracle"
	DBCockroach  DBType = "cockroach"
	DBMongoDB    DBType = "mongodb"
	DBRedis      DBType = "redis"
	DBCassandra  DBType = "cassandra"
	DBClickHouse DBType = "clickhouse"
	DBNeo4j      DBType = "neo4j"
	DBDynamoDB   DBType = "dynamodb"
)

type AuthType string

const (
	AuthPassword AuthType = "password"
	AuthSSH      AuthType = "ssh"
	AuthSSL      AuthType = "ssl"
	AuthNone     AuthType = "none"
)

type ConnectionConfig struct {
	ID        string            `yaml:"id"`
	Name      string            `yaml:"name"`
	Type      DBType            `yaml:"type"`
	Host      string            `yaml:"host"`
	Port      int               `yaml:"port"`
	Database  string            `yaml:"database"`
	Username  string            `yaml:"username"`
	Password  string            `yaml:"Password"`
	AuthType  AuthType          `yaml:"auth_type"`
	SSHConfig *SSHConfig        `yaml:"ssh,omitempty"`
	SSLConfig *SSLConfig        `yaml:"ssl,omitempty"`
	Extra     map[string]string `yaml:"extra,omitempty"`
	Timeout   time.Time         `yaml:"timeout"`
	MaxOpen   int               `yaml:"max_open"`
	MaxIdle   int               `yaml:"max_idle"`
	Color     string            `yaml:"color"`
	CreatedAt time.Time         `yaml:"created_at"`
	UpdatedAt time.Time         `yaml:"updated_at"`
}

type SSHConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	AuthType string `yaml:"auth_type"`
	Password string `yaml:"password"`
	KeyPath  string `yaml:"key_path,omitempty"`
}

type SSLConfig struct {
	Mode     string `yaml:"mode"`
	CAFile   string `yaml:"ca_file"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type ConnectionInfo struct {
	Config        *ConnectionConfig
	Connected     bool
	ConnectedAt   time.Time
	ServerVersion string
	DefaultSchema string
}
