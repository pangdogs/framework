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
	leaseId     etcdv3.LeaseID
}

// KeepAliveContinuous 节点持续保活
func (r *_EtcdRegistration) KeepAliveContinuous(ctx context.Context) (async.Future, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-r.registry.ctx.Done():
		return async.Future{}, errors.New("registry: registry is terminating")
	default:
	}

	if !r.registry.barrier.Join(1) {
		return async.Future{}, errors.New("registry: registry is terminating")
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-r.registry.ctx.Done():
		}
		cancel()
	}()

	keepAliveChan, err := r.registry.client.KeepAlive(ctx, r.leaseId)
	if err != nil {
		cancel()
		r.registry.barrier.Done()

		log.L(r.registry.svcCtx).Error("keep alive etcd lease failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].Id.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseId)),
			zap.Error(err))
		return async.Future{}, fmt.Errorf("registry: %w", err)
	}

	terminated := async.NewFutureVoid()

	go func() {
		defer func() {
			cancel()
			r.registry.barrier.Done()
		}()

		for range keepAliveChan {
			log.L(r.registry.svcCtx).Debug("keep alive etcd lease heartbeat ok",
				zap.String("service", r.serviceNode.Name),
				zap.String("node", r.serviceNode.Nodes[0].Id.String()),
				zap.String("key", r.nodeKey),
				zap.Int64("lease_id", int64(r.leaseId)))
		}

		log.L(r.registry.svcCtx).Debug("keep alive etcd lease heartbeat closed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].Id.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseId)))

		async.ReturnVoid(terminated)
	}()

	log.L(r.registry.svcCtx).Debug("keep alive etcd lease ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].Id.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseId)))
	return terminated.Out(), nil
}

// KeepAliveOnce 节点保活一次
func (r *_EtcdRegistration) KeepAliveOnce(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := r.registry.client.KeepAliveOnce(ctx, r.leaseId)
	if err != nil {
		log.L(r.registry.svcCtx).Error("keep alive etcd lease once failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].Id.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseId)),
			zap.Error(err))
		return fmt.Errorf("registry: %w", err)
	}

	log.L(r.registry.svcCtx).Debug("keep alive etcd lease once ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].Id.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseId)))
	return nil
}

// Deregister 注销服务节点
func (r *_EtcdRegistration) Deregister(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := r.registry.client.Revoke(ctx, r.leaseId)
	if err != nil {
		log.L(r.registry.svcCtx).Error("revoke etcd lease failed",
			zap.String("service", r.serviceNode.Name),
			zap.String("node", r.serviceNode.Nodes[0].Id.String()),
			zap.String("key", r.nodeKey),
			zap.Int64("lease_id", int64(r.leaseId)),
			zap.Error(err))
		return fmt.Errorf("registry: %w", err)
	}

	log.L(r.registry.svcCtx).Debug("deregister service node ok",
		zap.String("service", r.serviceNode.Name),
		zap.String("node", r.serviceNode.Nodes[0].Id.String()),
		zap.String("key", r.nodeKey),
		zap.Int64("lease_id", int64(r.leaseId)))
	return nil
}

func (r *_EtcdRegistry) registerNode(ctx context.Context, serviceName string, node *discovery.Node, ttl time.Duration) (discovery.IRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	nodeKey := r.newNodeKey(serviceName, node.Id)

	grantRsp, err := r.client.Grant(ctx, int64(math.Ceil(max(ttl.Seconds(), 3))))
	if err != nil {
		log.L(r.svcCtx).Error("grant new etcd lease failed",
			zap.String("service", serviceName),
			zap.String("node", node.Id.String()),
			zap.String("key", nodeKey),
			zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}
	leaseId := grantRsp.ID

	serviceNode := &discovery.Service{
		Name:  serviceName,
		Nodes: []discovery.Node{*node},
	}
	serviceNodeData := encodeService(serviceNode)

	rsp, err := r.client.Txn(ctx).
		If(etcdv3.Compare(etcdv3.Version(nodeKey), "=", 0)).
		Then(etcdv3.OpPut(nodeKey, serviceNodeData, etcdv3.WithLease(leaseId))).
		Commit()
	if err != nil {
		log.L(r.svcCtx).Error("put etcd key failed",
			zap.String("service", serviceName),
			zap.String("node", node.Id.String()),
			zap.String("key", nodeKey),
			zap.Int64("lease_id", int64(leaseId)),
			zap.Error(err))
		return nil, fmt.Errorf("registry: %w", err)
	}
	if !rsp.Succeeded {
		log.L(r.svcCtx).Error("put etcd key failed",
			zap.String("service", serviceName),
			zap.String("node", node.Id.String()),
			zap.String("key", nodeKey),
			zap.Int64("lease_id", int64(leaseId)),
			zap.Error(discovery.ErrDuplicateRegistration))
		return nil, discovery.ErrDuplicateRegistration
	}

	serviceNode.Revision = rsp.Header.Revision

	registration := &_EtcdRegistration{
		registry:    r,
		nodeKey:     nodeKey,
		serviceNode: serviceNode,
		leaseId:     leaseId,
	}

	log.L(r.svcCtx).Debug("register service node ok",
		zap.String("service", serviceName),
		zap.String("node", node.Id.String()),
		zap.String("key", nodeKey),
		zap.Int64("lease_id", int64(leaseId)))
	return registration, nil
}
