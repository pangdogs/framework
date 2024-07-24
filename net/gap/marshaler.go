package gap

import (
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
)

// Marshal 序列化
func Marshal[T MsgReader](msg T) (ret binaryutil.RecycleBytes, err error) {
	bs := binaryutil.MakeRecycleBytes(msg.Size())
	defer func() {
		if !bs.Equal(ret) {
			bs.Release()
		}
	}()

	if _, err := msg.Read(bs.Data()); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("marshal msg(%d) failed, %w", msg.MsgId(), err)
	}

	return bs, nil
}

// Unmarshal 反序列化
func Unmarshal[T MsgWriter](msg T, data []byte) error {
	if _, err := msg.Write(data); err != nil {
		return fmt.Errorf("unmarshal msg(%d) failed, %w", msg.MsgId(), err)
	}
	return nil
}
