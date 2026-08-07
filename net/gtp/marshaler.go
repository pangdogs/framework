/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package gtp

import (
	"fmt"

	"git.golaxy.org/framework/utils/binaryutil"
)

// Marshal 编码消息并返回池化字节缓冲区；调用方使用完后必须调用 Release。
func Marshal[T ReadableMsg](msg T) (binaryutil.Bytes, error) {
	bs := binaryutil.NewBytes(true, msg.Size())
	if _, err := binaryutil.CopyToBuff(bs.Payload(), msg); err != nil {
		bs.Release()
		return binaryutil.EmptyBytes, fmt.Errorf("%w: marshal msg(%d) failed, %w", ErrGTP, msg.MsgId(), err)
	}
	return bs, nil
}

// Unmarshal 从 data 解码消息；消息中的引用型字段可能直接引用 data。
func Unmarshal(msg Msg, data []byte) error {
	if _, err := msg.Write(data); err != nil {
		return fmt.Errorf("%w: unmarshal msg(%d) failed, %w", ErrGTP, msg.MsgId(), err)
	}
	return nil
}
