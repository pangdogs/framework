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

package etcd_dsync

import (
	"context"
	"crypto/tls"
	"fmt"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/net/netpath"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func newDSync(settings ...option.Setting[DSyncOptions]) dsync.IDistSync {
	return &_DistSync{
		options: option.Make(With.Default(), settings...),
	}
}

type _DistSync struct {
	svcCtx  service.Context
	options DSyncOptions
	client  *etcdv3.Client
}

// Init 初始化插件
func (s *_DistSync) Init(svcCtx service.Context) {
	log.Infof(svcCtx, "init addin %q", self.Name)

	s.svcCtx = svcCtx

	if s.options.EtcdClient == nil {
		cli, err := etcdv3.New(s.configure())
		if err != nil {
			log.Panicf(svcCtx, "new etcd client failed, %s", err)
		}
		s.client = cli
	} else {
		s.client = s.options.EtcdClient
	}

	for _, ep := range s.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(s.svcCtx, 3*time.Second)
			defer cancel()

			if _, err := s.client.Status(ctx, ep); err != nil {
				log.Panicf(s.svcCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}
}

// Shut 关闭插件
func (s *_DistSync) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)

	if s.options.EtcdClient == nil {
		if s.client != nil {
			s.client.Close()
		}
	}
}

// NewMutex returns a new distributed mutex with given name.
func (s *_DistSync) NewMutex(name string, settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return s.newMutex(name, option.Make(dsync.With.Default(), settings...))
}

// NewMutexf returns a new distributed mutex using a formatted string.
func (s *_DistSync) NewMutexf(format string, args ...any) func(settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return func(settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
		return s.NewMutex(fmt.Sprintf(format, args...), settings...)
	}
}

// NewMutexp returns a new distributed mutex using elements.
func (s *_DistSync) NewMutexp(elems ...string) func(settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
	return func(settings ...option.Setting[dsync.DistMutexOptions]) dsync.IDistMutex {
		return s.NewMutex(netpath.Join(s.GetSeparator(), elems...), settings...)
	}
}

// GetSeparator return name path separator.
func (s *_DistSync) GetSeparator() string {
	return "/"
}

func (s *_DistSync) configure() etcdv3.Config {
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
