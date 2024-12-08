package geerpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method // 方法本身，用于调用方法，例如 Foo.Sum
	ArgType   reflect.Type   // 参数的类型，用于判断参数是否正确
	ReplyType reflect.Type   // 返回的类型，用于判断返回值是否正确
	numCalls  uint64         // 方法调用次数，用于统计方法调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv 创建一个参数变量
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	// 参数可能是指针类型也可能是值类型
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem()) // 指针类型，则创建指针类型变量
	} else {
		argv = reflect.New(m.ArgType).Elem() // 值类型，则创建值类型变量
	}
	return argv
}

// newReplyv 创建一个传出参数变量
func (m *methodType) newReplyv() reflect.Value {
	// 传出参数必须是指针类型
	replyv := reflect.New(m.ReplyType.Elem())

	// 为什么需要对 map 和 slice 特殊处理？
	// reflect.New 创建的 map 和 slice 都是 nil，需要先初始化，再使用。
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	name   string                 // 变量类型名，用来打 log，例如字符串 "Foo"
	typ    reflect.Type           // 变量类型，通过变量的类型，可以直接获取变量的方法，例如 Foo 类型，可获取到方法 Sum
	rcvr   reflect.Value          // 变量的值，通过变量的值，可以调用变量的方法，例如 &foo，调用 Sum
	method map[string]*methodType //  map 储存变量的方法中所有可调用的方法
}

// newService 通过发射获取 rcvr 变量的类型及其方法
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)                  // 获取 rcvr 变量的值，例如 &foo
	s.name = reflect.Indirect(s.rcvr).Type().Name() // 获取 rcvr 变量的类型名，例如 "Foo"
	s.typ = reflect.TypeOf(rcvr)                    // 获取 rcvr 变量的类型，例如 Foo
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods() // 获取 rcvr 变量的方法列表，例如 Foo.Sum，并将其保存到 service 的方法 map 中
	return s
}

// registerMethods 获取 rcvr 变量的方法列表，例如 Foo.Sum，并将其保存到 service 的方法 map 中
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)

	for i := 0; i < s.typ.NumMethod(); i++ { // 遍历 rcvr 变量的方法列表
		method := s.typ.Method(i) // 获取 rcvr 变量，第 i 个方法
		mType := method.Type      // 获取 rcvr 变量的方法的类型

		// 检查方法的参数是否正确。rpc 的方法必须满足以下条件：
		// 1. 方法有 3 个参数，第 1 个参数是 rcvr 变量(相当于 python 的 self，java 的 this)，第 2 个参数是传入参数，第 3 个参数是传出参数，指针类型
		// 2. 第 2 个参数和第 3 个参数都是导出的类型
		// 3. 返回值只有一个，是 error 类型
		// 例如这样 func (t *T) MethodName(argType T1, replyType *T2) error
		if mType.NumIn() != 3 || mType.NumOut() != 1 { // 检查参数个数
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() { // 检测返回值个数和类型
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) { // 检查参数类型是否是导出的类型或者内置类型
			continue
		}

		// 检查完毕，将方法保存到 service 的方法 map 中
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

// isExportedOrBuiltinType 检查类型是否是导出的类型或者内置类型
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// call 调用 rcvr 变量的方法，例如 Foo.Sum
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	// 用反射的方式调用方法
	// 第一个参数是 rcvr 变量，例如 &foo，类似于 java 的 this，python 的 self
	// 第 2 个参数是传入参数，第 3 个参数是传出参数，指针类型
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
