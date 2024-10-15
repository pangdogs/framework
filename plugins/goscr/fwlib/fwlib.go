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

package fwlib

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}

//go:generate yaegi extract git.golaxy.org/core git.golaxy.org/core/define git.golaxy.org/core/ec git.golaxy.org/core/event git.golaxy.org/core/plugin git.golaxy.org/core/pt git.golaxy.org/core/runtime git.golaxy.org/core/service git.golaxy.org/core/utils  git.golaxy.org/core/utils/async git.golaxy.org/core/utils/exception git.golaxy.org/core/utils/generic git.golaxy.org/core/utils/iface git.golaxy.org/core/utils/meta git.golaxy.org/core/utils/option git.golaxy.org/core/utils/reinterpret git.golaxy.org/core/utils/types git.golaxy.org/core/utils/uid
//go:generate yaegi extract git.golaxy.org/framework git.golaxy.org/framework/net/gap git.golaxy.org/framework/net/gtp git.golaxy.org/framework/net/netpath git.golaxy.org/framework/net/gap/codec git.golaxy.org/framework/net/gap/variant git.golaxy.org/framework/net/gtp/codec git.golaxy.org/framework/net/gtp/method git.golaxy.org/framework/net/gtp/transport git.golaxy.org/framework/plugins/broker git.golaxy.org/framework/plugins/conf git.golaxy.org/framework/plugins/db git.golaxy.org/framework/plugins/dentq git.golaxy.org/framework/plugins/dentr git.golaxy.org/framework/plugins/discovery git.golaxy.org/framework/plugins/dsvc git.golaxy.org/framework/plugins/dsync git.golaxy.org/framework/plugins/gate git.golaxy.org/framework/plugins/goscr git.golaxy.org/framework/plugins/log git.golaxy.org/framework/plugins/router git.golaxy.org/framework/plugins/rpc git.golaxy.org/framework/plugins/rpcstack git.golaxy.org/framework/plugins/broker/nats_broker git.golaxy.org/framework/plugins/db/dbutil git.golaxy.org/framework/plugins/db/mongodb git.golaxy.org/framework/plugins/db/redisdb git.golaxy.org/framework/plugins/db/sqldb git.golaxy.org/framework/plugins/discovery/cache_discovery git.golaxy.org/framework/plugins/discovery/etcd_discovery git.golaxy.org/framework/plugins/discovery/redis_discovery git.golaxy.org/framework/plugins/dsync/etcd_dsync git.golaxy.org/framework/plugins/dsync/redis_dsync git.golaxy.org/framework/plugins/gate/cli git.golaxy.org/framework/plugins/log/console_log git.golaxy.org/framework/plugins/log/zap_log git.golaxy.org/framework/plugins/rpc/callpath git.golaxy.org/framework/plugins/rpc/rpcli git.golaxy.org/framework/plugins/rpc/rpcpcsr git.golaxy.org/framework/plugins/rpc/rpcutil git.golaxy.org/framework/utils/binaryutil git.golaxy.org/framework/utils/concurrent
