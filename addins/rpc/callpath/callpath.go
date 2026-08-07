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

package callpath

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"git.golaxy.org/core/utils/types"
	"git.golaxy.org/core/utils/uid"
)

// TargetKind 标识 RPC 调用路径指向的对象类型。
type TargetKind uint8

const (
	// Service 表示服务实例。
	Service TargetKind = 'S'
	// Runtime 表示实体所在的运行时。
	Runtime TargetKind = 'R'
	// Entity 表示实体或实体组件。
	Entity TargetKind = 'E'
	// Client 表示网关客户端脚本。
	Client TargetKind = 'C'
)

// CallPath 描述 RPC 的目标类型、目标 ID、脚本及方法。
type CallPath struct {
	// TargetKind 指定调用目标的类型。
	TargetKind TargetKind
	// ExcludeSrc 指示路由时是否排除调用来源。
	ExcludeSrc bool
	// Id 是 Runtime 或 Entity 目标的实体 ID，其他目标类型忽略此字段。
	Id uid.Id
	// Script 是服务插件、运行时插件、实体组件或客户端脚本的名称；为空时表示目标本身。
	Script string
	// Method 是要调用的方法名。
	Method string
}

// Encode 编码调用路径。short 为 true 时使用进程内缓存索引压缩脚本名和方法名。
func (cp CallPath) Encode(short bool) ([]byte, error) {
	var sb bytes.Buffer

	sb.WriteByte(byte(cp.TargetKind))
	sb.WriteByte(types.Bool2Int[uint8](short)<<0 + types.Bool2Int[uint8](cp.ExcludeSrc)<<1)

	switch cp.TargetKind {
	case Service, Client:
		break
	case Runtime, Entity:
		sb.WriteString(cp.Id.String())
		sb.WriteByte(0)
	default:
		return nil, fmt.Errorf("rpc: invalid call path target kind: %c", cp.TargetKind)
	}

	if short {
		var buff [4]byte
		binary.LittleEndian.PutUint32(buff[:], reduce(cp.Script, cp.Method))
		sb.Write(buff[:])
	} else {
		sb.WriteString(cp.Script)
		sb.WriteByte(0)
		sb.WriteString(cp.Method)
		sb.WriteByte(0)
	}

	return sb.Bytes(), nil
}

// String 返回便于诊断的调用路径文本。
func (cp CallPath) String() string {
	switch cp.TargetKind {
	case Service:
		return fmt.Sprintf("%c-%d>%s.%s", cp.TargetKind, types.Bool2Int[int](cp.ExcludeSrc), cp.Script, cp.Method)
	case Runtime:
		return fmt.Sprintf("%c-%d-%s>%s.%s", cp.TargetKind, types.Bool2Int[int](cp.ExcludeSrc), cp.Id, cp.Script, cp.Method)
	case Entity:
		return fmt.Sprintf("%c-%d-%s>%s.%s", cp.TargetKind, types.Bool2Int[int](cp.ExcludeSrc), cp.Id, cp.Script, cp.Method)
	case Client:
		return fmt.Sprintf("%c>%s.%s", cp.TargetKind, cp.Script, cp.Method)
	}
	return ""
}

// Parse 解码调用路径；压缩路径要求本进程已缓存对应的脚本和方法。
func Parse(data []byte) (CallPath, error) {
	if len(data) < 2 {
		return CallPath{}, errors.New("rpc: invalid call path data format")
	}

	var cp CallPath
	offset := 0

	cp.TargetKind = TargetKind(data[offset])
	offset++

	cp.ExcludeSrc = (uint8(data[offset]>>1) & 0x1) != 0
	short := (uint8(data[offset]>>0) & 0x1) != 0
	offset++

	readStr := func() (string, error) {
		if len(data) < offset+1 {
			return "", errors.New("rpc: invalid call path data format")
		}

		l := bytes.IndexByte(data[offset:], 0)
		if l < 0 {
			return "", errors.New("rpc: invalid call path data format")
		}
		end := offset + l

		s := string(data[offset:end])
		offset += l + 1

		return s, nil
	}

	switch cp.TargetKind {
	case Service, Client:
		break
	case Runtime, Entity:
		{
			str, err := readStr()
			if err != nil {
				return CallPath{}, err
			}
			cp.Id = uid.Id(str)
		}
	default:
		return CallPath{}, fmt.Errorf("rpc: invalid call path target kind: %c", cp.TargetKind)
	}

	if short {
		if len(data[offset:]) < 4 {
			return CallPath{}, errors.New("rpc: invalid call path data format")
		}

		cached := inflate(binary.LittleEndian.Uint32(data[offset:]))
		if cached == nil {
			return CallPath{}, errors.New("rpc: inflate cached index failed")
		}

		cp.Script = cached.Script
		cp.Method = cached.Method

		return cp, nil
	}

	{
		str, err := readStr()
		if err != nil {
			return CallPath{}, err
		}
		cp.Script = str
	}

	{
		str, err := readStr()
		if err != nil {
			return CallPath{}, err
		}
		cp.Method = str
	}

	return cp, nil
}
