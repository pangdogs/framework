package dbutil

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type DB interface {
	*gorm.DB | *mongo.Client | *redis.Client
}

type IMigrateDB[T DB] interface {
	MigrateDB(db T) error
}

func MigrateDB[T DB](db T, services ...any) error {
	for _, service := range services {
		migrateDB, ok := service.(IMigrateDB[T])
		if !ok {
			continue
		}
		if err := migrateDB.MigrateDB(db); err != nil {
			return err
		}
	}
	return nil
}
