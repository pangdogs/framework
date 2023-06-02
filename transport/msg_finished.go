package transport

import "kit.golaxy.org/plugins/transport/binaryutil"

const (
	Flag_EncryptOK Flag = 1 << (iota + Flag_Options) // 加密成功，在服务端发起的Finished消息携带
	Flag_AuthOK                                      // 鉴权成功，在服务端发起的Finished消息携带
	Flag_Continue                                    // 断线重连，在Finished消息中携带
)

// MsgFinished 握手结束，表示服务器已认可客户端
type MsgFinished struct {
	SendSeq uint32 // 发送消息序号
	RecvSeq uint32 // 接收消息序号
}

func (m *MsgFinished) Read(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)

	return bs.BytesRead(), nil
}

func (m *MsgFinished) Write(p []byte) (int, error) {
	bs := binaryutil.NewByteStream(p)

	return bs.BytesWritten(), nil
}

func (m *MsgFinished) Size() int {

}
