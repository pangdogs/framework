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

package dsync

import (
	"math/rand"
	"time"

	"git.golaxy.org/core"
	"git.golaxy.org/core/utils/exception"
	"git.golaxy.org/core/utils/generic"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
)

type (
	// RetryDelayFunc 根据已尝试次数计算下一次重试前的等待时间。
	RetryDelayFunc = generic.Func1[int, time.Duration]
	// GenUIDFunc 生成锁所有权唯一 ID。
	GenUIDFunc = generic.FuncPair0[string, error]
)

// DistMutexOptions 配置分布式锁的租约、重试及所有权标识策略。
type DistMutexOptions struct {
	Expiry         time.Duration  // Expiry 是锁租约的有效期。
	Tries          int            // Tries 是后端尝试获取锁的次数。
	RetryDelayFunc RetryDelayFunc // RetryDelayFunc 计算相邻重试间隔。
	DriftFactor    float64        // DriftFactor 是计算有效期时预留的时钟漂移比例。
	TimeoutFactor  float64        // TimeoutFactor 是单次后端操作占租约时长的比例。
	GenUIDFunc     GenUIDFunc     // GenUIDFunc 为新锁生成所有权 ID。
	UID            string         // UID 显式指定所有权 ID；部分后端不支持。
}

// With 提供分布式锁的 Option 构造方法。
var With _DistMutexOption

type _DistMutexOption struct{}

// Default 返回八秒租约、最多 32 次尝试及随机 50-250 毫秒重试间隔的默认设置。
func (_DistMutexOption) Default() option.Setting[DistMutexOptions] {
	defaultRetryDelayFunc := func(tries int) time.Duration {
		const (
			minRetryDelayMilliSec = 50
			maxRetryDelayMilliSec = 250
		)
		return time.Duration(rand.Intn(maxRetryDelayMilliSec-minRetryDelayMilliSec)+minRetryDelayMilliSec) * time.Millisecond
	}

	defaultGenValueFunc := func() (string, error) {
		return string(uid.New()), nil
	}

	return func(options *DistMutexOptions) {
		With.Expiry(8 * time.Second).Apply(options)
		With.Tries(32).Apply(options)
		With.RetryDelayFunc(defaultRetryDelayFunc).Apply(options)
		With.DriftFactor(0.01).Apply(options)
		With.TimeoutFactor(0.10).Apply(options)
		With.GenUIDFunc(defaultGenValueFunc).Apply(options)
		With.UID("").Apply(options)
	}
}

// Expiry 设置分布式锁租约有效期。
func (_DistMutexOption) Expiry(expiry time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Expiry = expiry
	}
}

// Tries 设置获取分布式锁的最大尝试次数。
func (_DistMutexOption) Tries(tries int) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Tries = tries
	}
}

// RetryDelay 设置固定的重试间隔。
func (_DistMutexOption) RetryDelay(delay time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.RetryDelayFunc = func(tries int) time.Duration {
			return delay
		}
	}
}

// RetryDelayFunc 设置自定义重试间隔函数，不得为 nil。
func (_DistMutexOption) RetryDelayFunc(fn RetryDelayFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			exception.Panicf("dsync: %w: option RetryDelayFunc can't be assigned to nil", core.ErrArgs)
		}
		options.RetryDelayFunc = fn
	}
}

// DriftFactor 设置计算有效期时预留的时钟漂移比例。
func (_DistMutexOption) DriftFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.DriftFactor = factor
	}
}

// TimeoutFactor 设置单次后端操作占租约时长的比例。
func (_DistMutexOption) TimeoutFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.TimeoutFactor = factor
	}
}

// GenUIDFunc 设置锁所有权 ID 生成函数，不得为 nil。
func (_DistMutexOption) GenUIDFunc(fn GenUIDFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			exception.Panicf("dsync: %w: option GenUIDFunc can't be assigned to nil", core.ErrArgs)
		}
		options.GenUIDFunc = fn
	}
}

// UID 显式指定锁所有权 ID，以便在支持的后端间转移所有权。
func (_DistMutexOption) UID(v string) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.UID = v
	}
}
