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
	// Register 注册类型
	Register(v Value)
	// Deregister 取消注册类型
	Deregister(typeId TypeId)
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
	VariantCreator().Register(new(Int))
	VariantCreator().Register(new(Int8))
	VariantCreator().Register(new(Int16))
	VariantCreator().Register(new(Int32))
	VariantCreator().Register(new(Int64))
	VariantCreator().Register(new(Uint))
	VariantCreator().Register(new(Uint8))
	VariantCreator().Register(new(Uint16))
	VariantCreator().Register(new(Uint32))
	VariantCreator().Register(new(Uint64))
	VariantCreator().Register(new(Float))
	VariantCreator().Register(new(Double))
	VariantCreator().Register(new(Byte))
	VariantCreator().Register(new(Bool))
	VariantCreator().Register(new(Bytes))
	VariantCreator().Register(new(String))
	VariantCreator().Register(&Null{})
	VariantCreator().Register(&Map{})
	VariantCreator().Register(&Array{})
	VariantCreator().Register(&Error{})
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

// Register 注册类型
func (c *_VariantCreator) Register(v Value) {
	c.variantTypeMap.Insert(v.Type(), reflect.TypeOf(v).Elem())
}

// Deregister 取消注册类型
func (c *_VariantCreator) Deregister(typeId TypeId) {
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
