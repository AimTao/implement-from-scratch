package geecache

// 为什么要抽象 ByteView 数据类型来表示缓存值？
/* 1. 在 lru 中定义的 entry.value  是 Value 接口类型，所有传入的数据的类型均需要实现这个接口，也就是实现 Len 函数，比较麻烦。
      []byte 可以支持任何数据类型的存储，并为 []byte 抽象为 ByteView 类型，并实现 lru.Value 接口。
   2. 保证 ByteView 是只读类型
      如何保证？
      1. b 是小写，外部无法直接读取。
      2. 只能通过 ByteSlice 和 String 方法获取 b 的数据
          1.ByteSlice 返回 slice 时，会拷贝一个副本再返回
          2.String 将 b 的数据强制转换为 String。（外界也无法直接修改 b 的值）
*/

type ByteView struct {
	b []byte // slice
}

func (byteView ByteView) Len() int { // ByteView 实现 lru.Value 接口
	return len(byteView.b)
}

// ByteSlice 为什么要返回拷贝？
// 防止缓存值被外部修改，这里直接返回拷贝
func (byteView ByteView) ByteSlice() []byte {
	return cloneBytes(byteView.b) // 为什么不直接使用 make，要封装 cloneBytes 函数？方便复用 cloneBytes。
}

// 为什么需要 cloneBytes ?
// 因为 []byte 是切片，传递时，不会深拷贝，传递的是视图，底层数据会被外界修改
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (byteView ByteView) String() string {
	return string(byteView.b)
}
