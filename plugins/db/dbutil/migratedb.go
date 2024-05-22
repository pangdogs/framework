package dbutil

type IMigrateDB interface {
	MigrateDB() error
}
