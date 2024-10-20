// Code generated by 'yaegi extract git.golaxy.org/core/utils/async'. DO NOT EDIT.

package fwlib

import (
	"git.golaxy.org/core/utils/async"
	"git.golaxy.org/core/utils/generic"
	"reflect"
)

func init() {
	Symbols["git.golaxy.org/core/utils/async/async"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ErrAsyncRetClosed": reflect.ValueOf(&async.ErrAsyncRetClosed).Elem(),
		"MakeAsyncRet":      reflect.ValueOf(&async.MakeAsyncRet).Elem(),
		"MakeRet":           reflect.ValueOf(&async.MakeRet).Elem(),
		"VoidRet":           reflect.ValueOf(&async.VoidRet).Elem(),

		// type definitions
		"Callee": reflect.ValueOf((*async.Callee)(nil)),
		"Caller": reflect.ValueOf((*async.Caller)(nil)),

		// interface wrapper definitions
		"_Callee": reflect.ValueOf((*_git_golaxy_org_core_utils_async_Callee)(nil)),
		"_Caller": reflect.ValueOf((*_git_golaxy_org_core_utils_async_Caller)(nil)),
	}
}

// _git_golaxy_org_core_utils_async_Callee is an interface wrapper for Callee type
type _git_golaxy_org_core_utils_async_Callee struct {
	IValue                interface{}
	WPushCall             func(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WPushCallDelegate     func(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WPushCallVoid         func(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any]
	WPushCallVoidDelegate func(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any]
}

func (W _git_golaxy_org_core_utils_async_Callee) PushCall(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WPushCall(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallDelegate(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WPushCallDelegate(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallVoid(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WPushCallVoid(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Callee) PushCallVoidDelegate(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WPushCallVoidDelegate(fun, args...)
}

// _git_golaxy_org_core_utils_async_Caller is an interface wrapper for Caller type
type _git_golaxy_org_core_utils_async_Caller struct {
	IValue            interface{}
	WCall             func(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WCallDelegate     func(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any]
	WCallVoid         func(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any]
	WCallVoidDelegate func(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any]
}

func (W _git_golaxy_org_core_utils_async_Caller) Call(fun generic.FuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WCall(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallDelegate(fun generic.DelegateFuncVar0[any, async.RetT[any]], args ...any) async.AsyncRetT[any] {
	return W.WCallDelegate(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallVoid(fun generic.ActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WCallVoid(fun, args...)
}
func (W _git_golaxy_org_core_utils_async_Caller) CallVoidDelegate(fun generic.DelegateActionVar0[any], args ...any) async.AsyncRetT[any] {
	return W.WCallVoidDelegate(fun, args...)
}
