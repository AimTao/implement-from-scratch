package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

// Option 定义 Option 结构体，封装了 MagicNumber 和 CodecType 字段，从 conn 中解析出 Option 的信息，表示 RPC 消息的编码方式
type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

// Server 定义 Server 结构体，封装了 Accept、ServeConn、serveCodec 方法
type Server struct{}

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

type request struct {
	h                *codec.Header
	argv, replyValue reflect.Value
}

// readRequest 读取请求：调用 readRequestHeader 方法读取请求头，调用 ReadBody 方法读取请求参数，返回 request 结构体
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := request{
		h: h,
	}

	req.argv = reflect.New(reflect.TypeOf("")) // ?
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read body error: ", err)
	}

	return &req, nil
}

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

// handleRequest 处理请求：构造请求响应信息，调用 sendResponse 方法发送响应
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Println("handleRequest: ", req.h, req.argv.Elem())
	req.replyValue = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq)) // 构造请求的响应信息
	server.sendResponse(cc, req.h, req.replyValue.Interface(), sending)        // 发送响应
}

func (server *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock() // 加锁，防止并发写
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response error: ", err)
	}
}

var DefaultServer = NewServer()

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}
