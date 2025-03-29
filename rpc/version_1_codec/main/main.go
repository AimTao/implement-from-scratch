package main

import (
	geerpc "GeeRPC"
	"GeeRPC/codec"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	addr := make(chan string)
	// 启动服务端
	go startServer(addr)

	// 以下是客户端的逻辑
	addrString := <-addr
	fmt.Println(addrString)
	conn, _ := net.Dial("tcp", addrString)
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)

	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)            // 解析 conn 的前 8 个字节，解析 Option 信息，储存到 DefaultOption JSON 中
	cc := codec.NewCodecFuncMap[geerpc.DefaultOption.CodecType](conn) // 根据 DefaultOption 中的 CodecType 选择对应的 codec

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}

		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))

		var replyHeader codec.Header
		_ = cc.ReadHeader(&replyHeader)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("main: reply: ", replyHeader, reply)
	}
}

func startServer(addr chan string) {
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error", err)
	}
	log.Println("startServer: start rpc server on", listen.Addr()) // ?
	addr <- listen.Addr().String()
	geerpc.Accept(listen)
}
