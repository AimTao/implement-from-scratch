package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"log"
	"net"
	"sync"
)

// Call 实例表示一次 RPC 调用请求
type Call struct {
	Seq           uint64      // 请求的序号
	ServiceMethod string      // 请求的方法名
	Args          interface{} // 请求的参数
	Reply         interface{} // 请求的响应信息
	Error         error
	Done          chan *Call // 当调用结束后，会通过 Done 通知调用者  // 这个写法有意思，channel 的类型是 *Call
}

func (call *Call) done() {
	call.Done <- call
}

// Client 表示一个 RPC 客户端，一个客户端可以完成多个请求（Call 实例）的发送和接收
// 管理连接、请求和响应，可同时被多个协程并发使用
// 提供 Dial 方法，用于建立连接；提供 Call 方法，用于发送请求并等待响应结果
type Client struct {
	cc       codec.Codec      // 消息编解码器，用于序列化请求和反序列化响应
	opt      *Option          // 客户端配置，比如编码方式和协议参数。
	sending  sync.Mutex       // 互斥锁，用于确保在同一时间只有一个请求被发送
	header   codec.Header     // 每个请求，共用这个同一消息头
	mu       sync.Mutex       // 互斥锁，保护 pending 和 shutdown 字段，防止并发读写
	seq      uint64           // 用于给每个请求分配一个编号，用于区分不同的请求。（每个请求间没有顺序要求）
	pending  map[uint64]*Call // 每个请求对应一个 Call 实例。未处理的请求会被保存在该字段中
	closing  bool             // 是否正在关闭连接
	shutdown bool             // 客户端是否已经关闭
}

var ErrShutdown = errors.New("client has been shut down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

// registerCall 方法用于注册一个 Call 实例，并返回该实例的序号。
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}

	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

// removeCall 方法用于从 pending 中移除一个 Call 实例，表示该请求已处理完成或已取消，并返回该实例。
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()

	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// terminateCalls 方法用于在客户端关闭时，终止所有未完成的调用，并通知调用者发生了错误
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) receive() {
	var err error
	for err == nil { // 这个写法，会在 err 不为 nil 时退出循环，所以只会处理一次错误
		// 读取请求头
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}

		// 根据 h.Seq 找到对应的 Call 实例，并从 pending 中移除。
		call := client.removeCall(h.Seq)

		// 三种处理响应的情况
		switch {
		case call == nil: // Call 实例不存在，（可能客户端已经取消请求，但服务器还是在响应请求），忽略该请求
			err = client.cc.ReadBody(nil)
		case h.Error != "": // Call 实例存在，但服务器返回了错误
			// 将错误信息写入 call.Error 中，调用 call.done() 通知调用方
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default: // Call 实例存在，服务器正常响应
			// 读取响应体，将响应信息写入 call.Reply 中，调用 call.done() 通知调用方
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// 如果发生错误（如连接断开），调用 terminateCalls 方法： 将所有未完成的调用（pending 中的所有调用）标记为错误状态。通知所有调用方，释放资源。
	client.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {

	// 用 JSON 数据通知服务器，客户端的编码方式
	// json.NewEncoder(conn) 创建一个 JSON Encoder 对象，Encode 方法将 opt 编码为 JSON 数据， JSON Encoder 对象将 Json 数据写入到 conn 中，也就是发给服务器
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}

	newCodecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if newCodecFunc == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	return newClientCodec(newCodecFunc(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string, opts ...*Option) (client *Client, err error) {

	// 默认使用 Gob 编码
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	// 真正拨号建立连接
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()

	// 返回 Client 实例
	return NewClient(conn, opt)
}

// send 发送请求到服务器
func (client *Client) send(call *Call) {
	// make sure that the client will send a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	// register this call.
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send the request
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		// call 可能为 nil
		// 比如由于网络或者某种错误，客户端在 receive() 中已经将该请求从 pending 中移除，此时 call 为 nil
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go 异步调用，不阻塞等待响应结果
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}

	// 为本次调用请求创建一个 Call 实例
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}

	// 将 Call 实例发送到客户端
	client.send(call)
	return call
}

// Call 同步调用，阻塞等待响应结果
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done // 读 Done 的数据，阻塞等待 receive() 处理完响应结果，对 Done 写值
	return call.Error
}
