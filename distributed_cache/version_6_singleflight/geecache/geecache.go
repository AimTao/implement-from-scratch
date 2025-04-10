package geecache

import (
	"fmt"
	"geecache/geecache/singleflight"
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
	peers     PeerPicker

	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
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

/*
                               是
(Get)接收 key --> 检查是否被缓存 -----> 返回缓存值 (1)
                |  否                              是
                |-----> (load)是否应当从远程节点获取 -----> (getFromPeer)与远程节点交互 --> 返回缓存值 (2)
                            |  否
                            |-----> (getLocally)调用`回调函数`，获取值，(populateCache)并添加到缓存 --> 返回缓存值 (3)
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

// load: 缓存不存在时，先在远程节点中查找，未果再调用 getLocally 获取数据
func (g *Group) load(key string) (ByteView, error) {
	val, err := g.loader.Do(key, func() (interface{}, error) {

		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok { // 找到 key 对应的远程节点。
				value, err := g.getFromPeer(peer, key) // 在远程节点中查找
				if err == nil {
					return value, nil
				}
				log.Println("[GeeCache] failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return val.(ByteView), nil
	}
	return ByteView{}, err
}

// 从远程节点获取数据，传入的是 PeerGetter 接口类型，只要实现了 Get 方法，就可以传入
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
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

// RegisterPeers 将实现了 PeerPicker 接口的变量注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

/*
使用一致性哈希选择节点        是                                    是
    |-----> 是否是远程节点 -----> HTTP 客户端访问远程节点 --> 成功？-----> 服务端返回返回值
                    |  否                                  ↓  否
                    |----------------------------> 回退到本地节点处理。
*/

// PeerGetter 获取到的远程节点，协助 Group 从远程节点获取缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 获取缓存值
}

// PeerPicker 协助 Group 通过 key 选择远程节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 根据 key 选择节点 PeerGetter
}