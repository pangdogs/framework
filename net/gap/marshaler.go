package gap

import (
	"fmt"
	"git.golaxy.org/framework/util/binaryutil"
)

func Marshal[T Msg](msg T) (binaryutil.RecycleBytes, error) {
	bs := binaryutil.MakeRecycleBytes(binaryutil.BytesPool.Get(msg.Size()))

	if _, err := msg.Read(bs.Data()); err != nil {
		bs.Release()
		return binaryutil.MakeNonRecycleBytes(nil), fmt.Errorf("marshal msg(%d) failed, %s", msg.MsgId(), err)
	}

	return bs, nil
}

func Unmarshal[T Msg](msg T, data []byte) error {
	if _, err := msg.Write(data); err != nil {
		return fmt.Errorf("unmarshal msg(%d) failed, %s", msg.MsgId(), err)
	}
	return nil
}
