// 定义一个 Gob 编码方式的结构体，对 encoding/gob 进行了封装，实现了 Codec 接口

package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer // 使用缓冲流来提高性能，防止阻塞
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// ReadHeader 封装了 encoding/gob 的 Decode 方法，从连接中读取 Header 信息
func (c *GobCodec) ReadHeader(header *Header) error {
	return c.dec.Decode(header)
}

// ReadBody 封装了 encoding/gob 的 Decode 方法，从连接中读取 Body 信息
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

// Write 封装了 encoding/gob 的 Encode 方法，将 Header 和 Body 信息写入连接中
func (c *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() { // 关闭前，先将 buf 中的数据写入连接中
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err = c.enc.Encode(header); err != nil {
		log.Panicln("rpc codec: gob error encoding header: ", err)
		return err
	}
	if err = c.enc.Encode(body); err != nil {
		log.Panicln("rpc codec: gob error encoding body: ", err)
		return err
	}
	return nil
}

// Close 封装了 io.ReadWriteCloser 的 Close 方法，关闭连接
// io.ReadWriteCloser 包含 io.Closer 类型，io.Closer 类型包含 Close 方法
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
