package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/**
一致性哈希解决什么问题？
    每个数据，都会准确的分配到一个节点上。访问时，根据该数据的hash值，确定该访问哪个节点。
    数据倾斜问题：服务器节点过少时，会导致数据无法均匀分配在各节点上。使用虚拟节点解决。
    删除节点或增加节点时，只需要调整该节点的数据。
*/

type Hash func(data []byte) uint32 // hash 函数类型

type Map struct { // 一致性哈希算法的主要数据结构
	hash     Hash           // 设置自定义一种哈希算法函数
	replicas int            // 虚拟节点的倍数
	keys     []int          // 哈希环，有序储存所有虚拟节点
	hashMap  map[int]string // 虚拟节点和真实节点的映射表，key 是虚拟节点的哈希值，value 是真实节点的名称
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil { // 默认使用 crc32.ChecksumIEEE 哈希算法
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 在哈希环 Map.keys 中加入真实节点对应的虚拟节点
// 节点名和虚拟节点名经过 hash 计算后得到哈希值，将哈希值加入 Map.keys 中，并排序。
func (m *Map) Add(keys ...string) {
	for _, key := range keys { // 对于每个真实节点，添加 m.replicas 个虚拟节点。
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 计算虚拟节点的 hash 值
			m.keys = append(m.keys, hash)                      // 虚拟节点加入环
			m.hashMap[hash] = key                              // 将虚拟节点和真实节点加入映射表
		}
	}
	sort.Ints(m.keys) // 对虚拟节点排序
}

// Del 在哈希环 Map.keys 中减少真实节点对应的虚拟节点
func (m *Map) Del(keys ...string) {

	// 删除真实节点与虚拟节点的映射
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			delete(m.hashMap, hash)
		}
	}

	// 重新构建哈希环（仅保留未被删除的哈希值）
	var newKeys []int
	for hash := range m.hashMap {
		newKeys = append(newKeys, hash)
	}
	sort.Ints(newKeys)
	m.keys = newKeys
}

// Get 获取 key 所在节点名
// key 经过 hash 计算后得到哈希值，在哈希环 Map.keys 上查找最接近的节点。
// 例如：key 的哈希值是 10000，哈希环上找到最接近的两个节点是 8000、11000，应该存在 8000 这个节点上。
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// 二分查找虚拟节点
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash // 找到第一个大于等于 hash 的节点，没有则返回 len(m.keys)
	})

	return m.hashMap[m.keys[index%len(m.keys)]]
}
