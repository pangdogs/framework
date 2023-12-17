package variant

import (
	"errors"
	"reflect"
	"sync"
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
	// Make 创建对象
	Make(typeId TypeId) (ValueReader, error)
	// NewReflected 创建反射对象指针
	NewReflected(typeId TypeId) (reflect.Value, error)
	// MakeReflected 创建反射对象
	MakeReflected(typeId TypeId) (reflect.Value, error)
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
		variantTypeMap: make(map[TypeId]reflect.Type),
	}
}

// _VariantCreator 可变类型对象构建器
type _VariantCreator struct {
	sync.RWMutex
	variantTypeMap map[TypeId]reflect.Type
}

// Register 注册类型
func (c *_VariantCreator) Register(v Value) {
	c.Lock()
	defer c.Unlock()

	c.variantTypeMap[v.Type()] = reflect.TypeOf(v).Elem()
}

// Deregister 取消注册类型
func (c *_VariantCreator) Deregister(typeId TypeId) {
	c.Lock()
	defer c.Unlock()

	delete(c.variantTypeMap, typeId)
}

// New 创建对象指针
func (c *_VariantCreator) New(typeId TypeId) (Value, error) {
	reflected, err := c.NewReflected(typeId)
	if err != nil {
		return nil, err
	}
	return reflected.Interface().(Value), nil
}

// Make 创建对象
func (c *_VariantCreator) Make(typeId TypeId) (ValueReader, error) {
	reflected, err := c.MakeReflected(typeId)
	if err != nil {
		return nil, err
	}
	return reflected.Interface().(ValueReader), nil
}

// NewReflected 创建反射对象指针
func (c *_VariantCreator) NewReflected(typeId TypeId) (reflect.Value, error) {
	c.RLock()
	defer c.RUnlock()

	rtype, ok := c.variantTypeMap[typeId]
	if !ok {
		return reflect.Value{}, ErrNotRegistered
	}

	return reflect.New(rtype), nil
}

// MakeReflected 创建反射对象
func (c *_VariantCreator) MakeReflected(typeId TypeId) (reflect.Value, error) {
	c.RLock()
	defer c.RUnlock()

	rtype, ok := c.variantTypeMap[typeId]
	if !ok {
		return reflect.Value{}, ErrNotRegistered
	}

	return reflect.Zero(rtype), nil
}
