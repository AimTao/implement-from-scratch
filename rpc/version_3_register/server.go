package geerpc

import (
	"encoding/json"
	"errors"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c

// Option 定义 Option 结构体，封装了 MagicNumber 和 CodecType 字段，从 conn 中解析出 Option 的信息，表示 RPC 消息的编码方式
type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

// Server 定义 Server 结构体，封装了 Accept、ServeConn、serveCodec 方法
type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

// Accept 处理连接：建立 socket 连接，使用 goroutine 处理连接
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept() // 建立 socket 连接
		if err != nil {
			log.Println("rpc server: accept error: ", err)
			return
		}
		go server.ServeConn(conn) // 使用 goroutine 处理连接
	}

}

// ServeConn 处理消息：解析出 Option 信息，根据 CodecType 选择对应的 codec，调用 serveCodec 方法处理剩下的消息
func (server *Server) ServeConn(conn io.ReadWriteCloser) {

	defer func() {
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil { // opt 是传出参数，读到 RPC 前面的 JSON 数据，这包含了 option 信息，表示 RPC 消息的编码方式
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

// serveCodec 处理请求：调用 readRequest 方法读取请求，调用 handleRequest 方法处理请求
func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		req, err := server.readRequest(cc)
		if err != nil {
			break
		}

		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// request 表示一次调用的所有信息
type request struct {
	h            *codec.Header // 请求头
	svc          *service      // 请求对应的服务，使用 svc.call 调用对应的方法
	mtype        *methodType   // 请求对应的方法，是 svc.call 的第一个参数
	argv, replyv reflect.Value // 方法的传入参数和传出参数，是 svc.call 的第二个和第三个参数
}

// readRequest 读取请求：调用 readRequestHeader 方法读取请求头，调用 ReadBody 方法读取请求参数，返回 request 结构体
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	// 读取请求头
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	// 初始化请求结构体
	req := &request{h: h}

	// 根据请求头中的 ServiceMethod 字段找到对应的服务和方法类型
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}

	// 创建传入参数和传出参数的反射对象
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// 检查请求传入参数的类型是否为指针类型，如果不是，则使用 Addr() 方法将 req.argv 转换为指针类型
	// 为什么？
	// 因为如果传入值是值类型，传入后，是值拷贝，不会修改传入变量的原值，所以需要使用 Addr() 获取地址后传入。
	argvi := req.argv.Interface() // 使用 interface() 方法将 req.argv 转换为 interface{} 类型，这样可以传入任意类型的参数
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	// ReadBody 方法会将请求参数解码到 argvi 中储存
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

// readRequestHeader 读取请求头：调用 ReadHeader 方法读取请求头，返回请求头结构体
func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc server: read header error: ", err)
		}
		return nil, err
	}
	return &h, nil
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

// handleRequest 处理请求：构造请求响应信息，调用 sendResponse 方法发送响应
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	err := req.svc.call(req.mtype, req.argv, req.replyv) // 调用

	if err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func (server *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock() // 加锁，防止并发写
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response error: ", err)
	}
}

var DefaultServer = NewServer()

func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)                                        // 为 rcvr 变量的类型创建 service 结构体
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup { // 调用 serviceMap.LoadOrStore 将 service 结构体保存到 map 中
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}
