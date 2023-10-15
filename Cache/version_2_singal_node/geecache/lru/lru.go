package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 最大的内存
	nBytes    int64                         // 已经被用作缓存的内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // 用 map 查找链表的节点
	OnEvicted func(key string, value Value) // 删除缓存节点的回调函数
}

// 链表节点的数据类型
type entry struct {
	key   string // 保存 key，删除链表缓存时，便于同步删除 map 中的 k/v 记录
	value Value  // 真正的缓存数据，允许 Value 是任何类型
}

type Value interface {
	Len() int // Len() 表示该数据的内存大小, 添加数据前,应该手动实现该数据类型的 Len 函数
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 增加缓存数据
func (c *Cache) Add(key string, value Value) {
	if _, ok := c.cache[key]; ok {
		c.Update(key, value)
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nBytes {
			c.RemoveOldest()
		}
	}
}

// RemoveOldest 删除链表尾部的缓存数据
// 用户并不需要调用，Add/Update时，可用空间不够，才会删除
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry) // 先获取链表节点的 Value 字段，类型是 any，并转化为 *entry 类型
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Update 更新缓存数据
func (c *Cache) Update(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		c.Add(key, value)
	}
}

// Get 获取缓存数据
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// Len 获取缓存数据的条数
func (c *Cache) Len() int {
	return c.ll.Len()
}
