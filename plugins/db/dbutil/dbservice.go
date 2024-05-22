package dbutil

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type DB interface {
	*gorm.DB | *mongo.Client | *redis.Client
}

type DBService[T DB] struct {
	DB T
}

func (s *DBService[T]) Init(db T) {
	s.DB = db
}
