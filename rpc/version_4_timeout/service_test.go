package geerpc

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}

type ServiceTemp int

// ServiceTemp 有一个方法 Timeout，该方法耗时2s
func (s ServiceTemp) Timeout(args int, reply *int) error {
	time.Sleep(time.Second * time.Duration(args))
	*reply = 0
	return nil
}

func TestClient_Call(t *testing.T) {
	t.Parallel()

	addrCh := make(chan string)
	go func(chan string) { // 启动一个服务器，监听 0 端口，注册 ServiceTemp 类型的对象，然后启动 Accept 方法，等待客户端连接
		var s ServiceTemp
		_ = Register(&s)
		l, _ := net.Listen("tcp", ":0")
		addrCh <- l.Addr().String()
		Accept(l)
	}(addrCh)
	addr := <-addrCh

	// 测试客户端处理调用超时的情况
	t.Run("client call timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr)

		ctx, _ := context.WithTimeout(context.Background(), time.Second) // 创建一个超时的 context，如果 1s 内没有返回结果，context.Done() 会传出信号 struct{}{}
		var reply int
		err := client.Call(ctx, "ServiceTemp.Timeout", 20, &reply) // 调用 ServiceTemp.Timeout 方法，Call 发生超时
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")
	})

	// 测试服务端处理超时的情况
	t.Run("server handle timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr, &Option{HandleTimeout: time.Second})
		var reply int
		err := client.Call(context.Background(), "ServiceTemp.Timeout", 20, &reply) // 调用 ServiceTemp.Timeout 方法，服务端处理超时
		_assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}
