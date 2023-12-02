package gap

import (
	"kit.golaxy.org/plugins/gap/variant"
	"kit.golaxy.org/plugins/util/binaryutil"
)

// MsgRPCRequest RPC请求
type MsgRPCRequest struct {
	CorrId    int64         // 关联Id，用于支持Future等异步模型
	EntityId  string        // 实体Id
	Component string        // 组件名
	Method    string        // 方法名
	Args      variant.Array // 参数列表
}

// Read implements io.Reader
func (m MsgRPCRequest) Read(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	if err := bs.WriteVarint(m.CorrId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.EntityId); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Component); err != nil {
		return bs.BytesWritten(), err
	}
	if err := bs.WriteString(m.Method); err != nil {
		return bs.BytesWritten(), err
	}
	n, err := m.Args.Read(bs.BuffUnwritten())
	if err != nil {
		return bs.BytesWritten() + n, nil
	}
	return bs.BytesWritten() + n, nil
}

// Write implements io.Writer
func (m *MsgRPCRequest) Write(p []byte) (int, error) {
	bs := binaryutil.NewBigEndianStream(p)
	corrId, err := bs.ReadVarint()
	if err != nil {
		return bs.BytesRead(), err
	}
	entityId, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	component, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	method, err := bs.ReadString()
	if err != nil {
		return bs.BytesRead(), err
	}
	var args variant.Array
	n, err := args.Write(bs.BuffUnread())
	if err != nil {
		return bs.BytesRead() + n, err
	}
	m.CorrId = corrId
	m.EntityId = entityId
	m.Component = component
	m.Method = method
	m.Args = args
	return bs.BytesRead() + n, nil
}

// Size 大小
func (m MsgRPCRequest) Size() int {
	return binaryutil.SizeofVarint(m.CorrId) +
		binaryutil.SizeofString(m.EntityId) +
		binaryutil.SizeofString(m.Component) +
		binaryutil.SizeofString(m.Method) +
		m.Args.Size()
}

// MsgId 消息Id
func (MsgRPCRequest) MsgId() MsgId {
	return MsgId_RPC_Request
}
