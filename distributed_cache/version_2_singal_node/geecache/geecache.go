package geecache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	groupMu sync.RWMutex
	groups  = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	groupMu.Lock()
	defer groupMu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	groupMu.RLock() // 为什么是读锁？
	g := groups[name]
	groupMu.Unlock()
	return g
}

/*
                          是
接收 key --> 检查是否被缓存 -----> 返回缓存值 (1)
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 (2)
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 (3)
Get 完成流程（1）
load + getLocally 完成流程(3)
*/

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 缓存没有命中，就从其他数据源中获取
	return g.load(key)
}

// 当缓存不存在，Get -> load -> getLocally 获取数据
// 为什么不直接使用 getLocally，还要封装一个 load ？
// load 会先从 远程分布式节点获取，获取不到才会使用 getLocally。此处属于预留设计。
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

// getLocally 调用用户的回调函数 g.getter.Get，获取数据，并使用 populateCache 添加数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 将获取到的数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
