package gap

// MsgPacket 消息包
type MsgPacket struct {
	Head MsgHead // 消息头
	Msg  Msg     // 消息
}

// Read implements io.Reader
func (mp MsgPacket) Read(p []byte) (int, error) {
	rn := 0

	n, err := mp.Head.Read(p)
	rn += n
	if err != nil {
		return rn, err
	}

	if mp.Msg == nil {
		return rn, nil
	}

	n, err = mp.Msg.Read(p[rn:])
	rn += n

	return rn, err
}

// Write implements io.Writer
func (mp *MsgPacket) Write(p []byte) (int, error) {
	wn := 0

	n, err := mp.Head.Write(p)
	wn += n
	if err != nil {
		return wn, err
	}

	if mp.Msg == nil {
		return wn, nil
	}

	n, err = mp.Msg.Write(p[wn:])
	wn += n

	return wn, err
}

// Size 大小
func (mp MsgPacket) Size() int {
	n := mp.Head.Size()

	if mp.Msg != nil {
		n += mp.Msg.Size()
	}

	return n
}