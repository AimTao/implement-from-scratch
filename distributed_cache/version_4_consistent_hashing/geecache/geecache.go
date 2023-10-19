package geecache

import (
    "fmt"
    "log"
    "sync"
)

// Getter 用于从各类的外部数据源获取数据
type Getter interface {
    Get(key string) ([]byte, error)
}

// GetterFunc 为了便于使用者传入匿名函数到 Getter 中
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
    return f(key)
}

// Group 负责与用户交互（获取缓存值），并拥有从外部数据源获取值并存储在缓存中的功能
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
    groupMu.RLock()
    g := groups[name]
    groupMu.RUnlock()
    return g
}

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

// load 缓存不存在，调用 getLocally 获取数据
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
