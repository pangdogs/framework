package db

const (
	MySQL      = "mysql"
	PostgreSQL = "postgresql"
	SQLServer  = "sqlserver"
	SQLite     = "sqlite"
	Redis      = "redis"
	MongoDB    = "mongodb"
)

type DBInfo struct {
	Tag     string `json:"tag,omitempty"`
	Type    string `json:"type,omitempty"`
	ConnStr string `json:"conn_str,omitempty"`
}
