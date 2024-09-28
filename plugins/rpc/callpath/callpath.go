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

type Category uint8

const (
	Service Category = 'S'
	Runtime Category = 'R'
	Entity  Category = 'E'
	Client  Category = 'C'
)

type CallPath struct {
	Category   Category
	ExcludeSrc bool
	Id         uid.Id
	Script     string
	Method     string
}

func (cp CallPath) Encode(short bool) ([]byte, error) {
	var sb bytes.Buffer

	sb.WriteByte(byte(cp.Category))
	sb.WriteByte(types.Bool2Int[uint8](short)<<0 + types.Bool2Int[uint8](cp.ExcludeSrc)<<1)

	switch cp.Category {
	case Service, Client:
		break
	case Runtime, Entity:
		sb.WriteString(cp.Id.String())
		sb.WriteByte(0)
	default:
		return nil, errors.New("rpc: invalid call path category")
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

func (cp CallPath) String() string {
	switch cp.Category {
	case Service:
		return fmt.Sprintf("%c[%d]>%s>%s", cp.Category, types.Bool2Int[int](cp.ExcludeSrc), cp.Script, cp.Method)
	case Runtime:
		return fmt.Sprintf("%c[%d]>%s>%s>%s", cp.Category, types.Bool2Int[int](cp.ExcludeSrc), cp.Id, cp.Script, cp.Method)
	case Entity:
		return fmt.Sprintf("%c[%d]>%s>%s>%s", cp.Category, types.Bool2Int[int](cp.ExcludeSrc), cp.Id, cp.Script, cp.Method)
	case Client:
		return fmt.Sprintf("%c>%s>%s", cp.Category, cp.Script, cp.Method)
	}
	return ""
}

func Parse(data []byte) (CallPath, error) {
	if len(data) < 2 {
		return CallPath{}, errors.New("rpc: invalid call path data format")
	}

	var cp CallPath
	offset := 0

	cp.Category = Category(data[offset])
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

	switch cp.Category {
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
		return CallPath{}, errors.New("rpc: invalid call path category")
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
