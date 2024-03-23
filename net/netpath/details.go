package netpath

// NodeDetails 节点地址信息
type NodeDetails struct {
	Domain             string // 主域
	BroadcastSubdomain string // 广播地址子域
	BalanceSubdomain   string // 负载均衡地址子域
	MulticastSubdomain string // 组播地址子域
	NodeSubdomain      string // 单播地址子域
	PathSeparator      string // 地址路径分隔符
}

func (d NodeDetails) InDomain(path string) bool {
	return InDir(d.PathSeparator, path, d.Domain)
}

func (d NodeDetails) SameDomain(path string) bool {
	return SameDir(d.PathSeparator, path, d.Domain)
}

func (d NodeDetails) InBroadcastSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.BroadcastSubdomain)
}

func (d NodeDetails) SameBroadcastSubdomain(path string) bool {
	return SameDir(d.PathSeparator, path, d.BroadcastSubdomain)
}

func (d NodeDetails) InBalanceSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.BalanceSubdomain)
}

func (d NodeDetails) SameBalanceSubdomain(path string) bool {
	return SameDir(d.PathSeparator, path, d.BalanceSubdomain)
}

func (d NodeDetails) InMulticastSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.MulticastSubdomain)
}

func (d NodeDetails) SameMulticastSubdomain(path string) bool {
	return SameDir(d.PathSeparator, path, d.MulticastSubdomain)
}

func (d NodeDetails) InNodeSubdomain(path string) bool {
	return InDir(d.PathSeparator, path, d.NodeSubdomain)
}

func (d NodeDetails) SameNodeSubdomain(path string) bool {
	return SameDir(d.PathSeparator, path, d.NodeSubdomain)
}
