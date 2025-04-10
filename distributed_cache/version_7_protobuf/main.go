package main

import (
	"flag"
	"fmt"
	"geecache/geecache"
	"log"
	"net/http"
)

// 模拟回源的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10,
		geecache.GetterFunc(func(key string) ([]byte, error) { // 回源函数
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// startCacheServer 每个节点启动一个 HTTP 服务器，用于接收请求，返回缓存值。
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr) // 初始化一个 HTTPPool 实例
	peers.Set(addrs...)                 // 将所有节点的地址加入到 HTTPPool 中
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers)) // 监听端口，让 NewHTTPPool 的 ServeHTTP 方法处理请求
}

// startAPIServer 启动一个 API 服务，用于测试。接收请求，根据 key 查找对应节点，访问节点，返回结果。
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"

	// 所有分布式节点的地址，包括本机
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()

	// 启动一个 API 服务，用于测试。接收请求，根据 key 查找对应节点，访问节点，返回结果。
	if api {
		startAPIServer(apiAddr, gee)
	} else {
		startCacheServer(addrMap[port], []string(addrs), gee)
	}
}