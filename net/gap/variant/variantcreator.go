package variant

import (
	"errors"
	"git.golaxy.org/framework/util/concurrent"
	"reflect"
)

var (
	ErrNotVariant    = errors.New("gap: not variant")            // 非可变类型
	ErrNotRegistered = errors.New("gap: variant not registered") // 类型未注册
)

// IVariantCreator 可变类型对象构建器接口
type IVariantCreator interface {
	// Declare 注册类型
	Declare(v Value)
	// Undeclare 取消注册类型
	Undeclare(typeId TypeId)
	// New 创建对象指针
	New(typeId TypeId) (Value, error)
	// NewReflected 创建反射对象指针
	NewReflected(typeId TypeId) (reflect.Value, error)
}

var variantCreator = _NewVariantCreator()

// VariantCreator 可变类型对象构建器
func VariantCreator() IVariantCreator {
	return variantCreator
}

func init() {
	VariantCreator().Declare(new(Int))
	VariantCreator().Declare(new(Int8))
	VariantCreator().Declare(new(Int16))
	VariantCreator().Declare(new(Int32))
	VariantCreator().Declare(new(Int64))
	VariantCreator().Declare(new(Uint))
	VariantCreator().Declare(new(Uint8))
	VariantCreator().Declare(new(Uint16))
	VariantCreator().Declare(new(Uint32))
	VariantCreator().Declare(new(Uint64))
	VariantCreator().Declare(new(Float))
	VariantCreator().Declare(new(Double))
	VariantCreator().Declare(new(Byte))
	VariantCreator().Declare(new(Bool))
	VariantCreator().Declare(new(Bytes))
	VariantCreator().Declare(new(String))
	VariantCreator().Declare(&Null{})
	VariantCreator().Declare(&Map{})
	VariantCreator().Declare(&Array{})
	VariantCreator().Declare(&Error{})
}

// _NewVariantCreator 创建可变类型对象构建器
func _NewVariantCreator() IVariantCreator {
	return &_VariantCreator{
		variantTypeMap: concurrent.MakeLockedMap[TypeId, reflect.Type](0),
	}
}

// _VariantCreator 可变类型对象构建器
type _VariantCreator struct {
	variantTypeMap concurrent.LockedMap[TypeId, reflect.Type]
}

// Declare 注册类型
func (c *_VariantCreator) Declare(v Value) {
	c.variantTypeMap.Insert(v.TypeId(), reflect.TypeOf(v).Elem())
}

// Undeclare 取消注册类型
func (c *_VariantCreator) Undeclare(typeId TypeId) {
	c.variantTypeMap.Delete(typeId)
}

// New 创建对象指针
func (c *_VariantCreator) New(typeId TypeId) (Value, error) {
	reflected, err := c.NewReflected(typeId)
	if err != nil {
		return nil, err
	}
	return reflected.Interface().(Value), nil
}

// NewReflected 创建反射对象指针
func (c *_VariantCreator) NewReflected(typeId TypeId) (reflect.Value, error) {
	rtype, ok := c.variantTypeMap.Get(typeId)
	if !ok {
		return reflect.Value{}, ErrNotRegistered
	}
	return reflect.New(rtype), nil
}
