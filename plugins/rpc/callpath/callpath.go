package callpath

import (
	"errors"
	"git.golaxy.org/core/util/uid"
	"strings"
)

const (
	Service = "S"
	Runtime = "R"
	Entity  = "E"
	Client  = "C"
)

var (
	Sep = byte('>')
)

type CallPath struct {
	Category  string
	EntityId  uid.Id
	Plugin    string
	Component string
	Method    string
}

func (cp CallPath) Encode() (string, error) {
	var sb strings.Builder

	switch cp.Category {
	case Service:
		sb.WriteString(cp.Category)
		sb.WriteByte(Sep)
		sb.WriteString(cp.Plugin)
		sb.WriteByte(Sep)
		sb.WriteString(cp.Method)

		return sb.String(), nil

	case Runtime:
		sb.WriteString(cp.Category)
		sb.WriteByte(Sep)
		sb.WriteString(cp.EntityId.String())
		sb.WriteByte(Sep)
		sb.WriteString(cp.Plugin)
		sb.WriteByte(Sep)
		sb.WriteString(cp.Method)

		return sb.String(), nil

	case Entity:
		sb.WriteString(cp.Category)
		sb.WriteByte(Sep)
		sb.WriteString(cp.EntityId.String())
		sb.WriteByte(Sep)
		sb.WriteString(cp.Component)
		sb.WriteByte(Sep)
		sb.WriteString(cp.Method)

		return sb.String(), nil

	case Client:
		sb.WriteString(cp.Category)
		sb.WriteByte(Sep)
		sb.WriteString(cp.EntityId.String())
		sb.WriteByte(Sep)
		sb.WriteString(cp.Method)

		return sb.String(), nil

	default:
		return "", errors.New("rpc: invalid action")
	}
}

func (cp CallPath) String() string {
	str, _ := cp.Encode()
	return str
}

func Parse(path string) (CallPath, error) {
	var cp CallPath

loop:
	for i := 0; ; i++ {
		idx := strings.IndexByte(path, Sep)
		if idx < 0 {
			if path == "" {
				break
			}
			idx = len(path)
		}
		field := path[:idx]

		switch i {
		case 0:
			cp.Category = field

			switch cp.Category {
			case Service, Runtime, Entity, Client:
			default:
				return CallPath{}, errors.New("rpc: invalid action")
			}

		case 1:
			switch cp.Category {
			case Service:
				cp.Plugin = field
			case Runtime, Entity, Client:
				cp.EntityId = uid.From(field)
			}

		case 2:
			switch cp.Category {
			case Service, Client:
				cp.Method = field
			case Runtime:
				cp.Plugin = field
			case Entity:
				cp.Component = field
			}

		case 3:
			switch cp.Category {
			case Runtime, Entity:
				cp.Method = field
			}

		default:
			break loop
		}

		if idx < len(path) {
			path = path[idx+1:]
			continue
		}

		break
	}

	return cp, nil
}
