package dentq

import (
	"crypto/tls"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/util/option"
	"git.golaxy.org/core/util/uid"
	"git.golaxy.org/framework/plugins/dserv"
	"git.golaxy.org/framework/plugins/log"
	"github.com/elliotchance/pie/v2"
	"github.com/josharian/intern"
	"github.com/muesli/cache2go"
	etcd_client "go.etcd.io/etcd/client/v3"
	"path"
	"strings"
	"sync"
	"time"
)

// DistEntity 分布式实体信息
type DistEntity struct {
	Id    uid.Id `json:"id"`    // 实体Id
	Nodes []Node `json:"nodes"` // 实体节点
}

// Node 实体节点信息
type Node struct {
	Service       string `json:"service"`        // 服务名称
	Id            string `json:"id"`             // 服务Id
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
		options: option.Make(Option{}.Default(), settings...),
	}
}

type _DistEntityQuerier struct {
	options  DistEntityQuerierOptions
	servCtx  service.Context
	distServ dserv.IDistService
	client   *etcd_client.Client
	wg       sync.WaitGroup
	cache    *cache2go.CacheTable
}

// InitSP 初始化服务插件
func (d *_DistEntityQuerier) InitSP(ctx service.Context) {
	log.Infof(ctx, "init plugin %q", plugin.Name)

	d.servCtx = ctx
	d.distServ = dserv.Using(d.servCtx)

	if d.options.EtcdClient == nil {
		cli, err := etcd_client.New(d.configure())
		if err != nil {
			log.Panicf(ctx, "new etcd client failed, %s", err)
		}
		d.client = cli
	} else {
		d.client = d.options.EtcdClient
	}

	for _, ep := range d.client.Endpoints() {
		if _, err := d.client.Status(ctx, ep); err != nil {
			log.Panicf(ctx, "status etcd %q failed, %s", ep, err)
		}
	}

	d.cache = cache2go.Cache(plugin.Name)

	d.wg.Add(1)
	go d.mainLoop()
}

// ShutSP 关闭服务插件
func (d *_DistEntityQuerier) ShutSP(ctx service.Context) {
	log.Infof(ctx, "shut plugin %q", plugin.Name)

	d.wg.Wait()

	if d.options.EtcdClient == nil {
		if d.client != nil {
			d.client.Close()
		}
	}
}

// GetDistEntity 查询分布式实体
func (d *_DistEntityQuerier) GetDistEntity(id uid.Id) (*DistEntity, bool) {
	item, err := d.cache.Value(id)
	if err == nil {
		return item.Data().(*DistEntity), true
	}

	rsp, err := d.client.Get(d.servCtx, path.Join(d.options.KeyPrefix, id.String()),
		etcd_client.WithPrefix(),
		etcd_client.WithSort(etcd_client.SortByModRevision, etcd_client.SortDescend),
		etcd_client.WithSerializable())
	if err != nil {
		return nil, false
	}

	entity := &DistEntity{
		Id: id,
	}

	for _, kv := range rsp.Kvs {
		subs := strings.Split(strings.TrimPrefix(string(kv.Key), d.options.KeyPrefix), "/")
		if len(subs) != 3 {
			continue
		}

		node := Node{
			Service: intern.String(subs[1]),
			Id:      intern.String(subs[2]),
		}
		node.BroadcastAddr = d.distServ.MakeBroadcastAddr(node.Service)
		node.BalanceAddr = d.distServ.MakeBalanceAddr(node.Service)
		node.RemoteAddr, _ = d.distServ.MakeNodeAddr(node.Id)

		entity.Nodes = append(entity.Nodes, node)
	}

	d.cache.NotFoundAdd(id, d.options.CacheExpiry, entity)

	return entity, true
}

func (d *_DistEntityQuerier) mainLoop() {
	defer d.wg.Done()

	log.Debug(d.servCtx, "watching distributed entities changes started")

retry:
	var watchChan etcd_client.WatchChan
	var uniqueList []string
	retryInterval := 3 * time.Second

	select {
	case <-d.servCtx.Done():
		goto end
	default:
	}

	watchChan = d.client.Watch(d.servCtx, d.options.KeyPrefix, etcd_client.WithPrefix())

	for watchRsp := range watchChan {
		if watchRsp.Canceled {
			log.Debugf(d.servCtx, "stop watch %q, retry it", d.options.KeyPrefix)
			time.Sleep(retryInterval)
			goto retry
		}
		if watchRsp.Err() != nil {
			log.Errorf(d.servCtx, "interrupt watch %q, %s, retry it", d.options.KeyPrefix, watchRsp.Err())
			time.Sleep(retryInterval)
			goto retry
		}

		uniqueList = uniqueList[:0]

		for _, event := range watchRsp.Events {
			subs := strings.Split(strings.TrimPrefix(string(event.Kv.Key), d.options.KeyPrefix), "/")
			if len(subs) != 3 {
				continue
			}

			id := subs[0]

			switch event.Type {
			case etcd_client.EventTypePut, etcd_client.EventTypeDelete:
				if pie.Contains(uniqueList, id) {
					continue
				}
				uniqueList = append(uniqueList, id)
				d.cache.Delete(id)
			default:
				continue
			}
		}
	}

end:
	log.Debug(d.servCtx, "watching distributed entities changes stopped")
}

func (d *_DistEntityQuerier) configure() etcd_client.Config {
	if d.options.EtcdConfig != nil {
		return *d.options.EtcdConfig
	}

	config := etcd_client.Config{
		Endpoints: d.options.CustomAddresses,
		Username:  d.options.CustomUsername,
		Password:  d.options.CustomPassword,
	}

	if d.options.CustomSecure || d.options.CustomTLSConfig != nil {
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
