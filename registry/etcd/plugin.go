package etcd

import (
	"github.com/galaxy-kit/galaxy-go/define"
	"github.com/galaxy-kit/plugins-go/registry"
)

// Plugin 定义本插件
var Plugin = define.DefinePlugin[registry.Registry, WithEtcdOption]().ServicePlugin(newRegistry)
