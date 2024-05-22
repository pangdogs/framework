package dbutil

type IMigrateDB interface {
	MigrateDB() error
}

func MigrateDB(services ...any) error {
	for _, service := range services {
		migrateDB, ok := service.(IMigrateDB)
		if !ok {
			continue
		}
		if err := migrateDB.MigrateDB(); err != nil {
			return err
		}
	}
	return nil
}
