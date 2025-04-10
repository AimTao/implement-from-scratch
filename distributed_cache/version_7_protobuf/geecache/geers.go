package geecache

import "geecache/geecache/geecachepb"

/*
使用一致性哈希选择节点        是                                    是
    |-----> 是否是远程节点 -----> HTTP 客户端访问远程节点 --> 成功？-----> 服务端返回返回值
                    |  否                                  ↓  否
                    |----------------------------> 回退到本地节点处理。
*/

// PeerGetter 获取到的远程节点，协助 Group 从远程节点获取缓存值
type PeerGetter interface {
	Get(in *geecachepb.Request, out *geecachepb.Response) error // 获取缓存值
}

// PeerPicker 协助 Group 通过 key 选择远程节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 根据 key 选择节点 PeerGetter
}
