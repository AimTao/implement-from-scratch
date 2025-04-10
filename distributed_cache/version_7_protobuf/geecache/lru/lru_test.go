package lru

import (
    "fmt"
    "testing"
)

// 对于数据类型，需要进行封装，因为无法在原生类型 string 中添加 Len 方法
type String string

func (s String) Len() int {
    return len(s)
}

// TestFunction 测试增删改查基本功能
func TestFunction(t *testing.T) {
    lru := New(int64(10), nil)

    lru.Add("k1", String("1234"))
    fmt.Println(lru.Get("k1"))

    lru.RemoveOldest()
    fmt.Println(lru.Get("k1"))

    lru.Update("k2", String("123"))
    fmt.Println(lru.Get("k2"))
}

// TestAutoRemoveOldest 测试当使用内存超过 maxBytes，是否会触发 “无用” 节点的删除
func TestAutoRemoveOldest(t *testing.T) {
    testData := []struct {
        key   string
        value Value
    }{
        {"k1", String("1234567890")},
        {"k2", String("234567890")},
        {"k3", String("34567890")},
    }

    lru := New(int64(10), nil)

    for _, test := range testData {
        lru.Add(test.key, test.value)
    }

    for _, test := range testData {
        fmt.Println(lru.Get(test.key))
    }
}

// TestCallbackOnEvicted 测试删除缓存时回调函数是否能被调用
func TestCallbackOnEvicted(t *testing.T) {
    keys := make([]string, 0)
    lru := New(10, func(key string, value Value) {
        keys = append(keys, key)
    })
    lru.Add("k1", String("k1"))
    lru.Add("k2", String("k2"))
    lru.Add("k3", String("k3"))
    lru.Add("k4", String("k4"))

    fmt.Println(keys)
}