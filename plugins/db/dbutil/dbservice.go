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
	db T
}

func (s *DBService[T]) bindDB(db T) {
	s.db = db
}

func (s *DBService[T]) DB() T {
	return s.db
}

type iBindDB[T DB] interface {
	bindDB(db T)
}

func BindDB[S iBindDB[T], T DB](service S, db T) S {
	service.bindDB(db)
	return service
}

type DBService2[T0, T1 DB] struct {
	db0 T0
	db1 T1
}

func (s *DBService2[T0, T1]) bindDB(db0 T0, db1 T1) {
	s.db0 = db0
	s.db1 = db1
}

func (s *DBService2[T0, T1]) DB0() T0 {
	return s.db0
}

func (s *DBService2[T0, T1]) DB1() T1 {
	return s.db1
}

type iBindDB2[T0, T1 DB] interface {
	bindDB(db0 T0, db1 T1)
}

func BindDB2[S iBindDB2[T0, T1], T0, T1 DB](service S, db0 T0, db1 T1) S {
	service.bindDB(db0, db1)
	return service
}

type DBService3[T0, T1, T2 DB] struct {
	db0 T0
	db1 T1
	db2 T2
}

func (s *DBService3[T0, T1, T2]) bindDB(db0 T0, db1 T1, db2 T2) {
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
}

func (s *DBService3[T0, T1, T2]) DB0() T0 {
	return s.db0
}

func (s *DBService3[T0, T1, T2]) DB1() T1 {
	return s.db1
}

func (s *DBService3[T0, T1, T2]) DB2() T2 {
	return s.db2
}

type iBindDB3[T0, T1, T2 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2)
}

func BindDB3[S iBindDB3[T0, T1, T2], T0, T1, T2 DB](service S, db0 T0, db1 T1, db2 T2) S {
	service.bindDB(db0, db1, db2)
	return service
}
