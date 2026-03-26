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

package dsync_etcd

import (
	"context"
	"crypto/tls"
	"time"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func newEtcdSync(settings ...option.Setting[EtcdSyncOptions]) dsync.IDistSync {
	return &_EtcdSync{
		options: option.New(With.Default(), settings...),
	}
}

type _EtcdSync struct {
	svcCtx  service.Context
	options EtcdSyncOptions
	client  *etcdv3.Client
}

// Init 初始化插件
func (s *_EtcdSync) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	s.svcCtx = svcCtx

	if s.options.EtcdClient == nil {
		cli, err := etcdv3.New(s.configure())
		if err != nil {
			log.L(svcCtx).Panic("new etcd client failed", log.JSON("config", s.configure()), zap.Error(err))
		}
		s.client = cli
	} else {
		s.client = s.options.EtcdClient
	}

	for _, ep := range s.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(svcCtx, 3*time.Second)
			defer cancel()

			if _, err := s.client.Status(ctx, ep); err != nil {
				log.L(svcCtx).Panic("status etcd failed", zap.Any("endpoint", ep), zap.Error(err))
			}
		}()
	}
}

// Shut 关闭插件
func (s *_EtcdSync) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))

	if s.options.EtcdClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex 创建分布式锁
func (s *_EtcdSync) NewMutex(name string, settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return s.newMutex(name, option.New(dsync.With.Default(), settings...))
}

// Separator 获取分隔符
func (s *_EtcdSync) Separator() string {
	return "/"
}

func (s *_EtcdSync) configure() etcdv3.Config {
	if s.options.EtcdConfig != nil {
		return *s.options.EtcdConfig
	}

	config := etcdv3.Config{
		Endpoints:   s.options.CustomAddresses,
		Username:    s.options.CustomUsername,
		Password:    s.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if s.options.CustomTLSConfig != nil {
		tlsConfig := s.options.CustomTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}
