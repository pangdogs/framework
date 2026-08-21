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

package discovery_etcd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/log"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type _EtcdRegistration struct {
	registry    *_EtcdRegistry
	nodeKey     string
	serviceNode *discovery.Service
	leaseID     etcdv3.LeaseID
}

// KeepAliveContinuous 持续刷新节点租约，直到 ctx、registry 或 ETCD 保活流结束。
func (r *_EtcdRegistration) KeepAliveContinuous(ctx context.Context) (async.Signal, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-r.registry.scope.Context().Done():
		return async.Signal{}, errors.New("registry: registry is terminating")
	default:
	}

	if !r.registry.barrier.Join(1) {
		return async.Signal{}, errors.New("registry: registry is terminating")
	}
	defer r.registry.barrier.Done()

	ctx, cancel := context.WithCancel(ctx)
	stopOwner := context.AfterFunc(r.registry.scope.Context(), cancel)
	stop := func() {
		stopOwner()
		cancel()
	}

	keepAliveChan, err := r.registry.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		stop()

		log.L(r.registry.svcCtx).Error("keep alive etcd lease failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].ID.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseID)),
			zap.Error(err))
		return async.Signal{}, fmt.Errorf("registry: %w", err)
	}

	stopped, stoppedSignal := async.NewSignal()

	future := async.SpawnVoid(r.registry.scope, func(context.Context) {
		defer stopped.Complete()
		defer stop()
		for range keepAliveChan {
			log.L(r.registry.svcCtx).Debug("keep alive etcd lease heartbeat ok",
				zap.String("service", r.serviceNode.Name),
				zap.String("node", r.serviceNode.Nodes[0].ID.String()),
				zap.String("key", r.nodeKey),
				zap.Int64("lease_id", int64(r.leaseID)))
		}

		log.L(r.registry.svcCtx).Debug("keep alive etcd lease heartbeat closed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].ID.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseID)))
	})
	future.OnComplete(func(ret async.Result) {
		if ret.Error != nil && !errors.Is(ret.Error, async.ErrScopeClosed) {
			log.L(r.registry.svcCtx).Error("registry keepalive task failed", zap.Error(ret.Error))
		}
	})

	if ret, ok := future.TryGet(); ok && errors.Is(ret.Error, async.ErrScopeClosed) {
		stop()
		for range keepAliveChan {
		}
		stopped.Complete()
		return async.Signal{}, errors.New("registry: registry is terminating")
	}

	log.L(r.registry.svcCtx).Debug("keep alive etcd lease ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].ID.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseID)))
	return stoppedSignal, nil
}

// KeepAliveOnce 立即向 ETCD 刷新一次节点租约。
func (r *_EtcdRegistration) KeepAliveOnce(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := r.registry.client.KeepAliveOnce(ctx, r.leaseID)
	if err != nil {
		log.L(r.registry.svcCtx).Error("keep alive etcd lease once failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].ID.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseID)),
			zap.Error(err))
		return fmt.Errorf("registry: %w", err)
	}

	log.L(r.registry.svcCtx).Debug("keep alive etcd lease once ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].ID.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseID)))
	return nil
}

// Deregister 撤销 ETCD 租约，从而注销服务节点及所有租约关联键。
func (r *_EtcdRegistration) Deregister(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := r.registry.client.Revoke(ctx, r.leaseID)
	if err != nil {
		log.L(r.registry.svcCtx).Error("revoke etcd lease failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].ID.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseID)),
			zap.Error(err))
		return fmt.Errorf("registry: %w", err)
	}

	log.L(r.registry.svcCtx).Debug("deregister service node ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].ID.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseID)))
	return nil
}

// registerNode 以不少于三秒的租约原子创建节点键；键已存在时返回重复注册错误。
func (r *_EtcdRegistry) registerNode(ctx context.Context, serviceName string, node *discovery.Node, ttl time.Duration) (discovery.IRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	nodeKey := r.newNodeKey(serviceName, node.ID)

	grantRsp, err := r.client.Grant(ctx, int64(math.Ceil(max(ttl.Seconds(), 3))))
	if err != nil {
		log.L(r.svcCtx).Error("grant etcd lease failed",
			zap.String("service", serviceName),
			zap.String("node", node.ID.String()),
			zap.String("key", nodeKey),
			zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}
	leaseID := grantRsp.ID

	serviceNode := &discovery.Service{
		Name:  serviceName,
		Nodes: []discovery.Node{*node},
	}
	serviceNodeData := encodeService(serviceNode)

	rsp, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(nodeKey), "=", 0)).
		Then(etcdv3.OpPut(nodeKey, serviceNodeData, etcdv3.WithLease(leaseID))).
		Commit()
	if err != nil {
		r.client.Revoke(context.Background(), leaseID)

		log.L(r.svcCtx).Error("put service node etcd key failed",
			zap.String("service", serviceName),
			zap.String("node", node.ID.String()),
			zap.String("key", nodeKey),
			zap.Int64("lease_id", int64(leaseID)),
			zap.Error(err))

		return nil, fmt.Errorf("registry: %w", err)
	}
	if !rsp.Succeeded {
		r.client.Revoke(context.Background(), leaseID)

		log.L(r.svcCtx).Error("put service node etcd key failed",
			zap.String("service", serviceName),
			zap.String("node", node.ID.String()),
			zap.String("key", nodeKey),
			zap.Int64("lease_id", int64(leaseID)),
			zap.Error(discovery.ErrDuplicateRegistration))

		return nil, discovery.ErrDuplicateRegistration
	}

	serviceNode.Revision = rsp.Header.Revision

	registration := &_EtcdRegistration{
		registry:    r,
		nodeKey:     nodeKey,
		serviceNode: serviceNode,
		leaseID:     leaseID,
	}

	log.L(r.svcCtx).Debug("register service node ok",
		zap.String("service", serviceName),
		zap.String("node", node.ID.String()),
		zap.String("key", nodeKey),
		zap.Int64("lease_id", int64(leaseID)))
	return registration, nil
}
