package netpath

// AddressDetails 地址信息
type AddressDetails struct {
	Domain             string // 主域
	BroadcastSubdomain string // 广播地址子域
	BalanceSubdomain   string // 负载均衡地址子域
	MulticastSubdomain string // 组播地址子域
	NodeSubdomain      string // 服务节点地址子域
	PathSeparator      string // 地址路径分隔符
}

func (ad AddressDetails) InDomain(path string) bool {
	return InDir(ad.PathSeparator, path, ad.Domain)
}

func (ad AddressDetails) InBroadcastSubdomain(path string) bool {
	return InDir(ad.PathSeparator, path, ad.BroadcastSubdomain)
}

func (ad AddressDetails) InBalanceSubdomain(path string) bool {
	return InDir(ad.PathSeparator, path, ad.BalanceSubdomain)
}

func (ad AddressDetails) InMulticastSubdomain(path string) bool {
	return InDir(ad.PathSeparator, path, ad.MulticastSubdomain)
}

func (ad AddressDetails) InNodeSubdomain(path string) bool {
	return InDir(ad.PathSeparator, path, ad.NodeSubdomain)
}
