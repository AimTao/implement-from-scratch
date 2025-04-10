package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.RWMutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {

	if g.m == nil { // 第一次不加锁，因为 g.m 的概率小
		g.mu.Lock() // 如果 g.m == nil，说明是第一次，需要加锁，避免其他并发也进来创建 g.m
		if g.m == nil {
			g.m = make(map[string]*call) // 懒加载
		}
		g.mu.Unlock()
	}

	g.mu.RLock()
	if c, ok := g.m[key]; ok {
		g.mu.RUnlock() // 读 g.m 后解锁
		c.wg.Wait()
		return c.val, c.err
	}
	g.mu.RUnlock() // 读 g.m 后解锁

	c := new(call)
	c.wg.Add(1) // 记录同步的事件数量，其他并发请求阻塞等待当前请求结束
	g.mu.Lock() // 写 g.m 前加锁
	g.m[key] = c
	g.mu.Unlock() // 写 g.m 后解锁

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 结束请求

	g.mu.Lock()      // 写 g.m 前加锁
	delete(g.m, key) // 更新 g.m
	g.mu.Unlock()    // 写 g.m 后解锁

	return c.val, c.err // 返回结果
}
