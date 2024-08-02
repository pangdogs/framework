/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package dbutil

import (
	"fmt"
	"git.golaxy.org/core"
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
	if db == nil {
		panic(fmt.Errorf("%s: db is nil", core.ErrArgs))
	}
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

type DBServiceSlice[T DB] struct {
	dbs []T
}

func (s *DBServiceSlice[T]) bindDB(dbs []T) {
	s.dbs = dbs
}

func (s *DBServiceSlice[T]) DB(idx int) T {
	if idx < 0 || idx >= len(s.dbs) {
		return nil
	}
	return s.dbs[idx]
}

type iBindDBSlice[T DB] interface {
	bindDB(dbs []T)
}

func BindDBSlice[S iBindDBSlice[T], T DB](service S, dbs ...T) S {
	service.bindDB(dbs)
	return service
}

type DBService2[T0, T1 DB] struct {
	db0 T0
	db1 T1
}

func (s *DBService2[T0, T1]) bindDB(db0 T0, db1 T1) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
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
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
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

type DBService4[T0, T1, T2, T3 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
}

func (s *DBService4[T0, T1, T2, T3]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
}

func (s *DBService4[T0, T1, T2, T3]) DB0() T0 {
	return s.db0
}

func (s *DBService4[T0, T1, T2, T3]) DB1() T1 {
	return s.db1
}

func (s *DBService4[T0, T1, T2, T3]) DB2() T2 {
	return s.db2
}

func (s *DBService4[T0, T1, T2, T3]) DB3() T3 {
	return s.db3
}

type iBindDB4[T0, T1, T2, T3 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3)
}

func BindDB4[S iBindDB4[T0, T1, T2, T3], T0, T1, T2, T3 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3) S {
	service.bindDB(db0, db1, db2, db3)
	return service
}

type DBService5[T0, T1, T2, T3, T4 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
}

func (s *DBService5[T0, T1, T2, T3, T4]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
}

func (s *DBService5[T0, T1, T2, T3, T4]) DB0() T0 {
	return s.db0
}

func (s *DBService5[T0, T1, T2, T3, T4]) DB1() T1 {
	return s.db1
}

func (s *DBService5[T0, T1, T2, T3, T4]) DB2() T2 {
	return s.db2
}

func (s *DBService5[T0, T1, T2, T3, T4]) DB3() T3 {
	return s.db3
}

func (s *DBService5[T0, T1, T2, T3, T4]) DB4() T4 {
	return s.db4
}

type iBindDB5[T0, T1, T2, T3, T4 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4)
}

func BindDB5[S iBindDB5[T0, T1, T2, T3, T4], T0, T1, T2, T3, T4 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4) S {
	service.bindDB(db0, db1, db2, db3, db4)
	return service
}

type DBService6[T0, T1, T2, T3, T4, T5 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
	db5 T5
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB0() T0 {
	return s.db0
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB1() T1 {
	return s.db1
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB2() T2 {
	return s.db2
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB3() T3 {
	return s.db3
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB4() T4 {
	return s.db4
}

func (s *DBService6[T0, T1, T2, T3, T4, T5]) DB5() T5 {
	return s.db5
}

type iBindDB6[T0, T1, T2, T3, T4, T5 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5)
}

func BindDB6[S iBindDB6[T0, T1, T2, T3, T4, T5], T0, T1, T2, T3, T4, T5 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5) S {
	service.bindDB(db0, db1, db2, db3, db4, db5)
	return service
}

type DBService7[T0, T1, T2, T3, T4, T5, T6 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
	db5 T5
	db6 T6
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB0() T0 {
	return s.db0
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB1() T1 {
	return s.db1
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB2() T2 {
	return s.db2
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB3() T3 {
	return s.db3
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB4() T4 {
	return s.db4
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB5() T5 {
	return s.db5
}

func (s *DBService7[T0, T1, T2, T3, T4, T5, T6]) DB6() T6 {
	return s.db6
}

type iBindDB7[T0, T1, T2, T3, T4, T5, T6 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6)
}

func BindDB7[S iBindDB7[T0, T1, T2, T3, T4, T5, T6], T0, T1, T2, T3, T4, T5, T6 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6)
	return service
}

type DBService8[T0, T1, T2, T3, T4, T5, T6, T7 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
	db5 T5
	db6 T6
	db7 T7
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB0() T0 {
	return s.db0
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB1() T1 {
	return s.db1
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB2() T2 {
	return s.db2
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB3() T3 {
	return s.db3
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB4() T4 {
	return s.db4
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB5() T5 {
	return s.db5
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB6() T6 {
	return s.db6
}

func (s *DBService8[T0, T1, T2, T3, T4, T5, T6, T7]) DB7() T7 {
	return s.db7
}

type iBindDB8[T0, T1, T2, T3, T4, T5, T6, T7 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7)
}

func BindDB8[S iBindDB8[T0, T1, T2, T3, T4, T5, T6, T7], T0, T1, T2, T3, T4, T5, T6, T7 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7)
	return service
}

type DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
	db5 T5
	db6 T6
	db7 T7
	db8 T8
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB0() T0 {
	return s.db0
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB1() T1 {
	return s.db1
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB2() T2 {
	return s.db2
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB3() T3 {
	return s.db3
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB4() T4 {
	return s.db4
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB5() T5 {
	return s.db5
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB6() T6 {
	return s.db6
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB7() T7 {
	return s.db7
}

func (s *DBService9[T0, T1, T2, T3, T4, T5, T6, T7, T8]) DB8() T8 {
	return s.db8
}

type iBindDB9[T0, T1, T2, T3, T4, T5, T6, T7, T8 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8)
}

func BindDB9[S iBindDB9[T0, T1, T2, T3, T4, T5, T6, T7, T8], T0, T1, T2, T3, T4, T5, T6, T7, T8 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8)
	return service
}

type DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 DB] struct {
	db0 T0
	db1 T1
	db2 T2
	db3 T3
	db4 T4
	db5 T5
	db6 T6
	db7 T7
	db8 T8
	db9 T9
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB0() T0 {
	return s.db0
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB1() T1 {
	return s.db1
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB2() T2 {
	return s.db2
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB3() T3 {
	return s.db3
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB4() T4 {
	return s.db4
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB5() T5 {
	return s.db5
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB6() T6 {
	return s.db6
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB7() T7 {
	return s.db7
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB8() T8 {
	return s.db8
}

func (s *DBService10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9]) DB9() T9 {
	return s.db9
}

type iBindDB10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9)
}

func BindDB10[S iBindDB10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9)
	return service
}

type DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB0() T0 {
	return s.db0
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB1() T1 {
	return s.db1
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB2() T2 {
	return s.db2
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB3() T3 {
	return s.db3
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB4() T4 {
	return s.db4
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB5() T5 {
	return s.db5
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB6() T6 {
	return s.db6
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB7() T7 {
	return s.db7
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB8() T8 {
	return s.db8
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB9() T9 {
	return s.db9
}

func (s *DBService11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) DB10() T10 {
	return s.db10
}

type iBindDB11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10)
}

func BindDB11[S iBindDB11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10)
	return service
}

type DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
	db11 T11
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	if db11 == nil {
		panic(fmt.Errorf("%s: db11 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
	s.db11 = db11
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB0() T0 {
	return s.db0
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB1() T1 {
	return s.db1
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB2() T2 {
	return s.db2
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB3() T3 {
	return s.db3
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB4() T4 {
	return s.db4
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB5() T5 {
	return s.db5
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB6() T6 {
	return s.db6
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB7() T7 {
	return s.db7
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB8() T8 {
	return s.db8
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB9() T9 {
	return s.db9
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB10() T10 {
	return s.db10
}

func (s *DBService12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) DB11() T11 {
	return s.db11
}

type iBindDB12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11)
}

func BindDB12[S iBindDB12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10, db11)
	return service
}

type DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
	db11 T11
	db12 T12
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	if db11 == nil {
		panic(fmt.Errorf("%s: db11 is nil", core.ErrArgs))
	}
	if db12 == nil {
		panic(fmt.Errorf("%s: db12 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
	s.db11 = db11
	s.db12 = db12
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB0() T0 {
	return s.db0
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB1() T1 {
	return s.db1
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB2() T2 {
	return s.db2
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB3() T3 {
	return s.db3
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB4() T4 {
	return s.db4
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB5() T5 {
	return s.db5
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB6() T6 {
	return s.db6
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB7() T7 {
	return s.db7
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB8() T8 {
	return s.db8
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB9() T9 {
	return s.db9
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB10() T10 {
	return s.db10
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB11() T11 {
	return s.db11
}

func (s *DBService13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) DB12() T12 {
	return s.db12
}

type iBindDB13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12)
}

func BindDB13[S iBindDB13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10, db11, db12)
	return service
}

type DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
	db11 T11
	db12 T12
	db13 T13
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	if db11 == nil {
		panic(fmt.Errorf("%s: db11 is nil", core.ErrArgs))
	}
	if db12 == nil {
		panic(fmt.Errorf("%s: db12 is nil", core.ErrArgs))
	}
	if db13 == nil {
		panic(fmt.Errorf("%s: db13 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
	s.db11 = db11
	s.db12 = db12
	s.db13 = db13
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB0() T0 {
	return s.db0
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB1() T1 {
	return s.db1
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB2() T2 {
	return s.db2
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB3() T3 {
	return s.db3
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB4() T4 {
	return s.db4
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB5() T5 {
	return s.db5
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB6() T6 {
	return s.db6
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB7() T7 {
	return s.db7
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB8() T8 {
	return s.db8
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB9() T9 {
	return s.db9
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB10() T10 {
	return s.db10
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB11() T11 {
	return s.db11
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB12() T12 {
	return s.db12
}

func (s *DBService14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) DB13() T13 {
	return s.db13
}

type iBindDB14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13)
}

func BindDB14[S iBindDB14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10, db11, db12, db13)
	return service
}

type DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
	db11 T11
	db12 T12
	db13 T13
	db14 T14
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	if db11 == nil {
		panic(fmt.Errorf("%s: db11 is nil", core.ErrArgs))
	}
	if db12 == nil {
		panic(fmt.Errorf("%s: db12 is nil", core.ErrArgs))
	}
	if db13 == nil {
		panic(fmt.Errorf("%s: db13 is nil", core.ErrArgs))
	}
	if db14 == nil {
		panic(fmt.Errorf("%s: db14 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
	s.db11 = db11
	s.db12 = db12
	s.db13 = db13
	s.db14 = db14
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB0() T0 {
	return s.db0
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB1() T1 {
	return s.db1
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB2() T2 {
	return s.db2
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB3() T3 {
	return s.db3
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB4() T4 {
	return s.db4
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB5() T5 {
	return s.db5
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB6() T6 {
	return s.db6
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB7() T7 {
	return s.db7
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB8() T8 {
	return s.db8
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB9() T9 {
	return s.db9
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB10() T10 {
	return s.db10
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB11() T11 {
	return s.db11
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB12() T12 {
	return s.db12
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB13() T13 {
	return s.db13
}

func (s *DBService15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) DB14() T14 {
	return s.db14
}

type iBindDB15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14)
}

func BindDB15[S iBindDB15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10, db11, db12, db13, db14)
	return service
}

type DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 DB] struct {
	db0  T0
	db1  T1
	db2  T2
	db3  T3
	db4  T4
	db5  T5
	db6  T6
	db7  T7
	db8  T8
	db9  T9
	db10 T10
	db11 T11
	db12 T12
	db13 T13
	db14 T14
	db15 T15
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14, db15 T15) {
	if db0 == nil {
		panic(fmt.Errorf("%s: db0 is nil", core.ErrArgs))
	}
	if db1 == nil {
		panic(fmt.Errorf("%s: db1 is nil", core.ErrArgs))
	}
	if db2 == nil {
		panic(fmt.Errorf("%s: db2 is nil", core.ErrArgs))
	}
	if db3 == nil {
		panic(fmt.Errorf("%s: db3 is nil", core.ErrArgs))
	}
	if db4 == nil {
		panic(fmt.Errorf("%s: db4 is nil", core.ErrArgs))
	}
	if db5 == nil {
		panic(fmt.Errorf("%s: db5 is nil", core.ErrArgs))
	}
	if db6 == nil {
		panic(fmt.Errorf("%s: db6 is nil", core.ErrArgs))
	}
	if db7 == nil {
		panic(fmt.Errorf("%s: db7 is nil", core.ErrArgs))
	}
	if db8 == nil {
		panic(fmt.Errorf("%s: db8 is nil", core.ErrArgs))
	}
	if db9 == nil {
		panic(fmt.Errorf("%s: db9 is nil", core.ErrArgs))
	}
	if db10 == nil {
		panic(fmt.Errorf("%s: db10 is nil", core.ErrArgs))
	}
	if db11 == nil {
		panic(fmt.Errorf("%s: db11 is nil", core.ErrArgs))
	}
	if db12 == nil {
		panic(fmt.Errorf("%s: db12 is nil", core.ErrArgs))
	}
	if db13 == nil {
		panic(fmt.Errorf("%s: db13 is nil", core.ErrArgs))
	}
	if db14 == nil {
		panic(fmt.Errorf("%s: db14 is nil", core.ErrArgs))
	}
	if db15 == nil {
		panic(fmt.Errorf("%s: db15 is nil", core.ErrArgs))
	}
	s.db0 = db0
	s.db1 = db1
	s.db2 = db2
	s.db3 = db3
	s.db4 = db4
	s.db5 = db5
	s.db6 = db6
	s.db7 = db7
	s.db8 = db8
	s.db9 = db9
	s.db10 = db10
	s.db11 = db11
	s.db12 = db12
	s.db13 = db13
	s.db14 = db14
	s.db15 = db15
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB0() T0 {
	return s.db0
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB1() T1 {
	return s.db1
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB2() T2 {
	return s.db2
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB3() T3 {
	return s.db3
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB4() T4 {
	return s.db4
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB5() T5 {
	return s.db5
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB6() T6 {
	return s.db6
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB7() T7 {
	return s.db7
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB8() T8 {
	return s.db8
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB9() T9 {
	return s.db9
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB10() T10 {
	return s.db10
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB11() T11 {
	return s.db11
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB12() T12 {
	return s.db12
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB13() T13 {
	return s.db13
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB14() T14 {
	return s.db14
}

func (s *DBService16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) DB15() T15 {
	return s.db15
}

type iBindDB16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 DB] interface {
	bindDB(db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14, db15 T15)
}

func BindDB16[S iBindDB16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15], T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 DB](service S, db0 T0, db1 T1, db2 T2, db3 T3, db4 T4, db5 T5, db6 T6, db7 T7, db8 T8, db9 T9, db10 T10, db11 T11, db12 T12, db13 T13, db14 T14, db15 T15) S {
	service.bindDB(db0, db1, db2, db3, db4, db5, db6, db7, db8, db9, db10, db11, db12, db13, db14, db15)
	return service
}
