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
	// RetryDelayFunc 重试延迟函数，用于决定两次重试之间需要等待的时间
	RetryDelayFunc = generic.Func1[int, time.Duration]
	// GenUIDFunc 生成唯一ID函数
	GenUIDFunc = generic.FuncPair0[string, error]
)

// DistMutexOptions 所有分布式锁选项
type DistMutexOptions struct {
	Expiry         time.Duration
	Tries          int
	RetryDelayFunc RetryDelayFunc
	DriftFactor    float64
	TimeoutFactor  float64
	GenUIDFunc     GenUIDFunc
	UID            string
}

var With _DistMutexOption

type _DistMutexOption struct{}

// Default 默认分布式锁选项
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

// Expiry 用于将分布式锁的过期时间设置为指定值
func (_DistMutexOption) Expiry(expiry time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Expiry = expiry
	}
}

// Tries 用于设置获取分布式锁的尝试次数
func (_DistMutexOption) Tries(tries int) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.Tries = tries
	}
}

// RetryDelay 用于设置两次重试之间需要等待的时间
func (_DistMutexOption) RetryDelay(delay time.Duration) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.RetryDelayFunc = func(tries int) time.Duration {
			return delay
		}
	}
}

// RetryDelayFunc 用于重写默认的重试延迟逻辑
func (_DistMutexOption) RetryDelayFunc(fn RetryDelayFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			exception.Panicf("dsync: %w: option RetryDelayFunc can't be assigned to nil", core.ErrArgs)
		}
		options.RetryDelayFunc = fn
	}
}

// DriftFactor 用于设置时钟漂移系数
func (_DistMutexOption) DriftFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.DriftFactor = factor
	}
}

// TimeoutFactor 用于设置超时系数
func (_DistMutexOption) TimeoutFactor(factor float64) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.TimeoutFactor = factor
	}
}

// GenUIDFunc 用于重写默认的唯一ID生成逻辑
func (_DistMutexOption) GenUIDFunc(fn GenUIDFunc) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		if fn == nil {
			exception.Panicf("dsync: %w: option GenUIDFunc can't be assigned to nil", core.ErrArgs)
		}
		options.GenUIDFunc = fn
	}
}

// UID 用于显式指定分布式锁的唯一ID，以便在无需重新加锁的情况下转移锁的所有权
func (_DistMutexOption) UID(v string) option.Setting[DistMutexOptions] {
	return func(options *DistMutexOptions) {
		options.UID = v
	}
}
