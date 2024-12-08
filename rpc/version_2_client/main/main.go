package main

import (
	"fmt"
	"geerpc"
	"geerpc/codec"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	log.SetFlags(0) // 设置日志的输出格式，不输出时间戳和文件名等信息

	addr := make(chan string)
	go startServer(addr)

	// 提前定义使用的编码方式，默认使用 Gob 编码
	opt := &geerpc.Option{
		CodecType: codec.GobType,
	}

	// 拨号建立连接
	client, _ := geerpc.Dial("tcp", <-addr, opt) // 主要需要实现的是 Dial 方法，该方法会返回一个 Client 实例
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)

	// 异步地发送请求，并打印响应结果
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ { // i 表示第几次发送，不是发送请求的 seq，seq 是在发送请求时自动生成的
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("geerpc req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil { // 主要需要实现的是 Call 方法，该方法会发送请求并等待响应结果
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d, reply: %s\n", i, reply)
		}(i)
	}
	wg.Wait()
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
