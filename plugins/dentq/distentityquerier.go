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

//go:generate go run k8s.io/code-generator/cmd/deepcopy-gen .
package dentq

import (
	"context"
	"crypto/tls"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/framework/utils/concurrent"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"path"
	"strings"
	"sync"
	"time"
	"unique"
)

// DistEntity 分布式实体信息
// +k8s:deepcopy-gen=true
type DistEntity struct {
	Id       uid.Id `json:"id"`       // 实体Id
	Nodes    []Node `json:"nodes"`    // 实体节点
	Revision int64  `json:"revision"` // 数据版本号
}

// Node 实体节点信息
// +k8s:deepcopy-gen=true
type Node struct {
	Service       string `json:"service"`        // 服务名称
	Id            uid.Id `json:"id"`             // 服务Id
	BroadcastAddr string `json:"broadcast_addr"` // 服务广播地址
	BalanceAddr   string `json:"balance_addr"`   // 服务负载均衡地址
	RemoteAddr    string `json:"remote_addr"`    // 远端服务节点地址
}

// IDistEntityQuerier 分布式实体信息查询器
type IDistEntityQuerier interface {
	// GetDistEntity 查询分布式实体
	GetDistEntity(id uid.Id) (*DistEntity, bool)
}

func newDistEntityQuerier(settings ...option.Setting[DistEntityQuerierOptions]) IDistEntityQuerier {
	return &_DistEntityQuerier{
		options: option.Make(With.Default(), settings...),
	}
}

type _DistEntityQuerier struct {
	svcCtx   service.Context
	options  DistEntityQuerierOptions
	distServ dserv.IDistService
	client   *etcdv3.Client
	wg       sync.WaitGroup
	cache    *concurrent.Cache[uid.Id, *DistEntity]
}

// InitSP 初始化服务插件
func (d *_DistEntityQuerier) InitSP(svcCtx service.Context) {
	log.Infof(svcCtx, "init plugin %q", self.Name)

	d.svcCtx = svcCtx
	d.distServ = dserv.Using(d.svcCtx)

	if d.options.EtcdClient == nil {
		cli, err := etcdv3.New(d.configure())
		if err != nil {
			log.Panicf(svcCtx, "new etcd client failed, %s", err)
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(d.svcCtx, 3*time.Second)
			defer cancel()

			if _, err := d.client.Status(ctx, ep); err != nil {
				log.Panicf(d.svcCtx, "status etcd %q failed, %s", ep, err)
			}
		}()
	}

	d.cache = concurrent.NewCache[uid.Id, *DistEntity]()
	d.cache.AutoClean(d.svcCtx, 30*time.Second, 256)

	d.wg.Add(1)
	go d.mainLoop()
}

// ShutSP 关闭服务插件
func (d *_DistEntityQuerier) ShutSP(svcCtx service.Context) {
	log.Infof(svcCtx, "shut plugin %q", self.Name)

	d.wg.Wait()

	if d.options.EtcdClient == nil {
		if d.client != nil {
			d.client.Close()
		}
	}
}

// GetDistEntity 查询分布式实体
func (d *_DistEntityQuerier) GetDistEntity(id uid.Id) (*DistEntity, bool) {
	entity, ok := d.cache.Get(id)
	if ok {
		return entity, true
	}

	rsp, err := d.client.Get(d.svcCtx, path.Join(d.options.KeyPrefix, id.String()),
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend),
		etcdv3.WithIgnoreValue())
	if err != nil || len(rsp.Kvs) <= 0 {
		return nil, false
	}

	entity = &DistEntity{
		Id:       id,
		Nodes:    make([]Node, 0, len(rsp.Kvs)),
		Revision: rsp.Header.Revision,
	}

	details := d.distServ.GetNodeDetails()

	for _, kv := range rsp.Kvs {
		subs := strings.Split(strings.TrimPrefix(string(kv.Key), d.options.KeyPrefix), "/")
		if len(subs) != 3 {
			continue
		}

		node := Node{
			Service: unique.Make(subs[1]).Value(),
			Id:      uid.From(unique.Make(subs[2]).Value()),
		}
		node.BroadcastAddr = details.MakeBroadcastAddr(node.Service)
		node.BalanceAddr = details.MakeBalanceAddr(node.Service)
		node.RemoteAddr, _ = details.MakeNodeAddr(node.Id)

		entity.Nodes = append(entity.Nodes, node)
	}

	return d.cache.Set(id, entity, entity.Revision, d.options.CacheExpiry), true
}

func (d *_DistEntityQuerier) mainLoop() {
	defer d.wg.Done()

	log.Debug(d.svcCtx, "watching distributed entities changes started")

retry:
	var watchChan etcdv3.WatchChan
	retryInterval := 3 * time.Second

	select {
	case <-d.svcCtx.Done():
		goto end
	default:
	}

	watchChan = d.client.Watch(d.svcCtx, d.options.KeyPrefix, etcdv3.WithPrefix(), etcdv3.WithIgnoreValue())

	for watchRsp := range watchChan {
		if watchRsp.Canceled {
			log.Debugf(d.svcCtx, "stop watch %q, retry it", d.options.KeyPrefix)
			time.Sleep(retryInterval)
			goto retry
		}
		if watchRsp.Err() != nil {
			log.Errorf(d.svcCtx, "interrupt watch %q, %s, retry it", d.options.KeyPrefix, watchRsp.Err())
			time.Sleep(retryInterval)
			goto retry
		}

		for _, event := range watchRsp.Events {
			subs := strings.Split(strings.TrimPrefix(string(event.Kv.Key), d.options.KeyPrefix), "/")
			if len(subs) != 3 {
				continue
			}

			entityId := uid.From(subs[0])

			switch event.Type {
			case etcdv3.EventTypePut, etcdv3.EventTypeDelete:
				d.cache.Del(entityId, watchRsp.Header.Revision)
			}
		}
	}

end:
	log.Debug(d.svcCtx, "watching distributed entities changes stopped")
}

func (d *_DistEntityQuerier) configure() etcdv3.Config {
	if d.options.EtcdConfig != nil {
		return *d.options.EtcdConfig
	}

	config := etcdv3.Config{
		Endpoints:   d.options.CustomAddresses,
		Username:    d.options.CustomUsername,
		Password:    d.options.CustomPassword,
		DialTimeout: 3 * time.Second,
	}

	if d.options.CustomTLSConfig != nil {
		tlsConfig := d.options.CustomTLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		config.TLS = tlsConfig
	}

	return config
}
