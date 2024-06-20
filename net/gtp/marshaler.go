package gtp

import (
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
)

func Marshal[T MsgReader](msg T) (binaryutil.RecycleBytes, error) {
	bs := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msg.Size()))

	if _, err := msg.Read(bs.Data()); err != nil {
		bs.Release()
		return binaryutil.NilRecycleBytes, fmt.Errorf("marshal msg(%d) failed, %w", msg.MsgId(), err)
	}

	return bs, nil
}

func Unmarshal[T MsgWriter](msg T, data []byte) error {
	if _, err := msg.Write(data); err != nil {
		return fmt.Errorf("unmarshal msg(%d) failed, %w", msg.MsgId(), err)
	}
	return nil
}
