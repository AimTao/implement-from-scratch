package geecache

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestGetter(t *testing.T) {
	var g Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	bytes, _ := g.Get("key")
	fmt.Printf("%c", bytes)
}

// 外部数据源
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db)) // 记录加入缓存的次数

	gee := NewGroup("scores", 1024, GetterFunc(
		func(key string) ([]byte, error) { // 缓存未命中时，从其他数据源获取数据的函数
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 从缓存中查看各个 key，测试缓存从外部数据源获取数据的能力
	for k, v := range db {
		view, err := gee.Get(k)
		if err != nil {
			panic(err)
		}
		fmt.Println(k, v, view.String())
	}

	// 对于数据源中不存在的数据的处理
	view, err := gee.Get("unknown")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(view.String())
}

func TestHTTP(t *testing.T) {
	NewGroup("scores", 1024, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}

/*
curl http://localhost:9999/_geecache/scores/Tom
curl http://localhost:9999/_geecache/scores/xxx

log:
2023/10/16 17:40:24 [Server localhost:9999] GET /_geecache/scores/Tom
2023/10/16 17:40:24 [SlowDB] search key Tom
2023/10/16 17:40:51 [Server localhost:9999] GET /_geecache/scores/567
2023/10/16 17:40:51 [SlowDB] search key 567
*/
