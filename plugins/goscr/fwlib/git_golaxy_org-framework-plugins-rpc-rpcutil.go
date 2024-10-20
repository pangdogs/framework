// Code generated by 'yaegi extract git.golaxy.org/framework/plugins/rpc/rpcutil'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/framework/plugins/rpc/rpcutil"
	"go/constant"
	"go/token"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/framework/plugins/rpc/rpcutil/rpcutil"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CliRPCPermValidator":    reflect.ValueOf(rpcutil.CliRPCPermValidator),
		"ConcurrentProxyEntity":  reflect.ValueOf(rpcutil.ConcurrentProxyEntity),
		"ConcurrentProxyGroup":   reflect.ValueOf(rpcutil.ConcurrentProxyGroup),
		"ConcurrentProxyRuntime": reflect.ValueOf(rpcutil.ConcurrentProxyRuntime),
		"NoComp":                 reflect.ValueOf(constant.MakeFromLiteral("\"\"", token.STRING, 0)),
		"NoPlugin":               reflect.ValueOf(constant.MakeFromLiteral("\"\"", token.STRING, 0)),
		"ProxyEntity":            reflect.ValueOf(rpcutil.ProxyEntity),
		"ProxyGroup":             reflect.ValueOf(rpcutil.ProxyGroup),
		"ProxyRuntime":           reflect.ValueOf(rpcutil.ProxyRuntime),
		"ProxyService":           reflect.ValueOf(rpcutil.ProxyService),

		// type definitions
		"EntityProxied":  reflect.ValueOf((*rpcutil.EntityProxied)(nil)),
		"GroupProxied":   reflect.ValueOf((*rpcutil.GroupProxied)(nil)),
		"RuntimeProxied": reflect.ValueOf((*rpcutil.RuntimeProxied)(nil)),
		"ServiceProxied": reflect.ValueOf((*rpcutil.ServiceProxied)(nil)),
	}
}
