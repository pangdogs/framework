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

package gap

import (
	"fmt"
	"git.golaxy.org/framework/utils/binaryutil"
)

// Marshal 序列化
func Marshal[T ReadableMsg](msg T) (ret binaryutil.RecycleBytes, err error) {
	bs := binaryutil.MakeRecycleBytes(msg.Size())
	defer func() {
		if !bs.Equal(ret) {
			bs.Release()
		}
	}()

	if _, err := binaryutil.CopyToBuff(bs.Data(), msg); err != nil {
		return binaryutil.NilRecycleBytes, fmt.Errorf("%w: marshal msg(%d) failed, %w", ErrGAP, msg.MsgId(), err)
	}

	return bs, nil
}

// Unmarshal 反序列化
func Unmarshal(msg Msg, data []byte) error {
	if _, err := msg.Write(data); err != nil {
		return fmt.Errorf("%w: unmarshal msg(%d) failed, %w", ErrGAP, msg.MsgId(), err)
	}
	return nil
}
