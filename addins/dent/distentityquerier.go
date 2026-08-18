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

package dent

import (
	"context"
	"crypto/tls"
	"path"
	"strings"
	"sync"
	"time"
	"unique"

	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/log"
	"github.com/dgraph-io/ristretto/v2"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// DistEntity 描述一个全局实体当前所在的服务节点集合。
type DistEntity struct {
	ID       uid.ID `json:"id"`       // ID 是实体 ID。
	Nodes    []Node `json:"nodes"`    // Nodes 是当前发布该实体的节点。
	Revision int64  `json:"revision"` // Revision 是查询时的 ETCD 全局修订号。
}

// Node 描述一个发布了分布式实体的服务节点及其消息地址。
type Node struct {
	Service       string `json:"service"`        // Service 是服务名称。
	ID            uid.ID `json:"id"`             // ID 是服务节点 ID。
	BroadcastAddr string `json:"broadcast_addr"` // BroadcastAddr 是该服务的广播地址。
	BalanceAddr   string `json:"balance_addr"`   // BalanceAddr 是该服务的负载均衡地址。
	RemoteAddr    string `json:"remote_addr"`    // RemoteAddr 是该节点的单播地址。
}

// IDistEntityQuerier 查询全局实体当前可达的服务节点。
type IDistEntityQuerier interface {
	// GetDistEntity 查询 id 对应的分布式实体信息。
	// 未找到或查询失败时返回 false；返回值可能来自缓存，不应由调用方修改。
	GetDistEntity(id uid.ID) (*DistEntity, bool)
}

func newDistEntityQuerier(settings ...option.Setting[DistEntityQuerierOptions]) IDistEntityQuerier {
	return &_DistEntityQuerier{
		options: option.New(With.Querier.Default(), settings...),
	}
}

type _DistEntityQuerier struct {
	svcCtx    service.Context
	ctx       context.Context
	terminate context.CancelFunc
	wg        sync.WaitGroup
	options   DistEntityQuerierOptions
	dsvc      dsvc.IDistService
	client    *etcdv3.Client
	cache     *ristretto.Cache[uid.ID, *DistEntity]
}

// Init 建立或复用 ETCD 客户端，检查端点状态，创建查询缓存并启动实体变更监听。
func (d *_DistEntityQuerier) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", QuerierAddIn.Name))

	d.svcCtx = svcCtx
	d.ctx, d.terminate = context.WithCancel(context.Background())

	d.dsvc = dsvc.AddIn.Require(svcCtx)

	if d.options.EtcdClient == nil {
		cli, err := etcdv3.New(d.configure())
		if err != nil {
			log.L(svcCtx).Panic("new etcd client failed", log.JSON("config", d.configure()), zap.Error(err))
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		func() {
			ctx, cancel := context.WithTimeout(svcCtx, 3*time.Second)
			defer cancel()

			if _, err := d.client.Status(ctx, ep); err != nil {
				log.L(svcCtx).Panic("status etcd failed", zap.Any("endpoint", ep), zap.Error(err))
			}
		}()
	}

	cache, err := ristretto.NewCache[uid.ID, *DistEntity](&ristretto.Config[uid.ID, *DistEntity]{
		NumCounters:        d.options.CacheNumCounters,
		MaxCost:            d.options.CacheMaxCost,
		BufferItems:        d.options.CacheBufferItems,
		ShouldUpdate:       func(cur, prev *DistEntity) bool { return cur.Revision > prev.Revision },
		IgnoreInternalCost: true,
	})
	if err != nil {
		log.L(svcCtx).Panic("new cache failed", zap.Error(err))
	}
	d.cache = cache

	d.wg.Add(1)
	go d.watchingForEntitiesChanges()
}

// Shut 停止实体变更监听并等待退出，关闭本地缓存；仅关闭由本 add-in 创建的 ETCD 客户端。
func (d *_DistEntityQuerier) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", QuerierAddIn.Name))

	d.terminate()
	d.wg.Wait()
	d.cache.Close()

	if d.options.EtcdClient == nil {
		if d.client != nil {
			d.client.Close()
		}
	}
}

// GetDistEntity 优先读取缓存，未命中时查询 ETCD 并组装节点地址。
// 未找到或查询失败时返回 false；返回对象可能是共享缓存值，调用方不得修改。
func (d *_DistEntityQuerier) GetDistEntity(id uid.ID) (*DistEntity, bool) {
	entity, ok := d.cache.Get(id)
	if ok {
		return entity, true
	}

	entityKey := path.Join(d.options.KeyPrefix, id.String())

	rsp, err := d.client.Get(d.svcCtx, entityKey+"/",
		etcdv3.WithKeysOnly(),
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByModRevision, etcdv3.SortDescend))
	if err != nil {
		log.L(d.svcCtx).Error("get distributed entity etcd key failed", zap.String("key", entityKey), zap.Error(err))
		return nil, false
	}
	if len(rsp.Kvs) <= 0 {
		return nil, false
	}

	entity = &DistEntity{
		ID:       id,
		Nodes:    make([]Node, 0, len(rsp.Kvs)),
		Revision: rsp.Header.Revision,
	}

	details := d.dsvc.NodeDetails()

	for _, kv := range rsp.Kvs {
		_, serviceName, nodeID, ok := d.parseEntityKey(string(kv.Key))
		if !ok {
			log.L(d.svcCtx).Error("invalid distributed entity key", zap.String("key", string(kv.Key)))
			continue
		}

		node := Node{
			Service: unique.Make(serviceName).Value(),
			ID:      uid.From(unique.Make(nodeID.String()).Value()),
		}
		node.BroadcastAddr = details.MakeBroadcastAddr(node.Service)
		node.BalanceAddr = details.MakeBalanceAddr(node.Service)
		node.RemoteAddr, _ = details.MakeNodeAddr(node.ID)

		entity.Nodes = append(entity.Nodes, node)
	}

	if d.cache.SetWithTTL(id, entity, 1, d.options.CacheTTL) {
		log.L(d.svcCtx).Debug("add distributed entity cache", zap.Any("id", entity.ID))
	}

	return entity, true
}

func (d *_DistEntityQuerier) watchingForEntitiesChanges() {
	defer d.wg.Done()

	log.L(d.svcCtx).Debug("watching for distributed entities changes started", zap.String("key", d.options.KeyPrefix))

	for watchRsp := range d.client.Watch(d.ctx, d.options.KeyPrefix, etcdv3.WithPrefix()) {
		if watchRsp.Canceled {
			log.L(d.svcCtx).Debug("watching etcd key canceled", zap.String("key", d.options.KeyPrefix), zap.Error(watchRsp.Err()))
			break
		}
		if watchRsp.Err() != nil {
			log.L(d.svcCtx).Panic("watching etcd key unexpectedly interrupted", zap.String("key", d.options.KeyPrefix), zap.Error(watchRsp.Err()))
		}

		for _, event := range watchRsp.Events {
			entityID, _, _, ok := d.parseEntityKey(string(event.Kv.Key))
			if !ok {
				log.L(d.svcCtx).Error("invalid distributed entity key", zap.String("key", string(event.Kv.Key)))
				continue
			}

			switch event.Type {
			case etcdv3.EventTypePut, etcdv3.EventTypeDelete:
				d.cache.Del(entityID)
				log.L(d.svcCtx).Debug("delete distributed entity cache", zap.Any("id", entityID))
			}
		}
	}

	log.L(d.svcCtx).Debug("watching for distributed entities changes stopped", zap.String("key", d.options.KeyPrefix))
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

func (d *_DistEntityQuerier) parseEntityKey(key string) (entityID uid.ID, serviceName string, nodeID uid.ID, ok bool) {
	subs := strings.Split(strings.TrimPrefix(key, d.options.KeyPrefix), "/")
	if len(subs) != 3 {
		return
	}

	entityID = uid.From(subs[0])
	serviceName = subs[1]
	nodeID = uid.From(subs[2])

	return entityID, serviceName, nodeID, true
}
