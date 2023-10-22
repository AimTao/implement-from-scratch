package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32 // hash 函数类型

type Map struct { // 一致性哈希算法的主要数据结构
	hash     Hash           // 设置自定义一种哈希算法函数
	replicas int            // 虚拟节点的倍数
	keys     []int          // 哈希环
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

// Add 添加真实节点
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

// Get 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// 二分查找虚拟节点
	index := sort.Search(len(m.keys), func(i int) bool { // 这个函数的意义
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[index%len(m.keys)]]
}
