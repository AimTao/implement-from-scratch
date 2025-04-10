package consistenthash

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

// 测试一致性哈希中哈希环的增加节点、查询节点的功能
func TestConsistentHash(t *testing.T) {

	// 初始化一致性 Map 时，需要传入自定义的哈希函数
	// 这里为了便于观察，传入的哈希函数，不进行哈希计算，直接返回节点名。
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	hash.Add("6", "4", "2") // 在一致性哈希的哈希环中，加入 节点6、节点4、节点2

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	// 在哈希环中，找到 k 对应的节点，判断节点是否和 正确答案v 一致。
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	hash.Add("8") // 在哈希环中增加新节点8

	testCases["27"] = "8" // 因为加入了新节点，经计算，key = 27 将从节点2 迁移到 节点8。

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}

// 测试虚拟节点个数，对于数据倾斜问题的影响
func TestConsistentHash2(t *testing.T) {
	hash := New(6, nil) // 传入 nil，则使用默认的哈希函数 crc32.ChecksumIEEE

	hash.Add("1", "2", "3", "4", "5", "6")
	m := make(map[int]int)

	num := 0

	for i := 0; i < 10000000; i++ {
		node := hash.Get(strconv.Itoa(i + rand.Int()))
		index, _ := strconv.Atoi(node)
		m[index]++ // 统计每个节点的个数
	}

	for key, value := range m {
		fmt.Printf("节点 %d 上的数据共 %d 个。\n", key, value)
		num += value
	}

	/* 对于 6个节点的情况，当虚拟节点为真实节点值的 6 倍及6倍以上时，分部还算均匀。

	节点 5 上的数据共 1618900 个。
	节点 3 上的数据共 2251837 个。
	节点 4 上的数据共 2293730 个。
	节点 6 上的数据共 1466977 个。
	节点 2 上的数据共 1278352 个。
	节点 1 上的数据共 1090204 个。
	*/
}
