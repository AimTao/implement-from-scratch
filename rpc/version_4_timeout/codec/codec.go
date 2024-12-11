// 定义一个抽象的 Codec 接口，Codec 接口包含了对消息体的编码和解码操作
// 具体的编码方式（如 Gob、Json）会来实现该接口

package codec

import "io"

type Header struct {
	ServiceMethod string // 服务名和方法名
	Seq           uint64 // 请求序号
	Error         string // 客户端置为空，服务端如何出现错误，将错误信息写入Error
}

// Codec 接口：对消息体进行编码，比如 Gob、Json 会实现该接口，代表两种编码方式
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// NewCodecFunc Codec 接口类型的构造函数
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

// NewCodecFuncMap 通过 type 来选择对应的 codec 的构造函数
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec // 赋值为 Gob codec 的构造函数
	//NewCodecFuncMap[JsonType] = NewJsonCodec  // 赋值为 Json codec 的构造函数, 暂不实现
}
