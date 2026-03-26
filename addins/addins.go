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

package addins

import (
	"git.golaxy.org/framework/addins/broker"
	"git.golaxy.org/framework/addins/broker/broker_nats"
	"git.golaxy.org/framework/addins/conf"
	"git.golaxy.org/framework/addins/db/mongodb"
	"git.golaxy.org/framework/addins/db/redisdb"
	"git.golaxy.org/framework/addins/db/sqldb"
	"git.golaxy.org/framework/addins/dent"
	"git.golaxy.org/framework/addins/discovery"
	"git.golaxy.org/framework/addins/discovery/discovery_etcd"
	"git.golaxy.org/framework/addins/dsvc"
	"git.golaxy.org/framework/addins/dsync"
	"git.golaxy.org/framework/addins/dsync/dsync_etcd"
	"git.golaxy.org/framework/addins/dsync/dsync_redis"
	"git.golaxy.org/framework/addins/gate"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/router"
	"git.golaxy.org/framework/addins/rpc"
	"git.golaxy.org/framework/addins/rpcstack"
)

var (
	Broker        = broker.AddIn
	BrokerNats    = broker_nats.AddIn
	Conf          = conf.AddIn
	MongoDB       = mongodb.AddIn
	RedisDB       = redisdb.AddIn
	SQLDB         = sqldb.AddIn
	Dentq         = dent.QuerierAddIn
	Dentr         = dent.RegistryAddIn
	Discovery     = discovery.AddIn
	DiscoveryEtcd = discovery_etcd.AddIn
	Dsvc          = dsvc.AddIn
	Dsync         = dsync.AddIn
	DsyncEtcd     = dsync_etcd.AddIn
	DsyncRedis    = dsync_redis.AddIn
	Gate          = gate.AddIn
	Log           = log.AddIn
	Router        = router.AddIn
	RPC           = rpc.AddIn
	RPCStack      = rpcstack.AddIn
)
