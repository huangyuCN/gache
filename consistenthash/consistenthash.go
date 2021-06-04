/**
consistenthash 一致性哈希
*/

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 将哈希值映射成一个非负整数
type Hash func(data []byte) uint32

//Map 一致性哈希的主体结构
type Map struct {
	hash     Hash           //哈希函数
	replicas int            //虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

// New 构造函数。运行自定义哈希算法。默认使用crc32.ChecksumIEEE 算法
func New(replicas int, f Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     f,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//Add 添加节点/机器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

//Get 获取节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	//顺时针找到第一个匹配的节点下标
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//从 m.keys 中获取到对应的哈希值。如果 idx == len(m.keys)，
	//说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	total := len(m.keys)
	rem := idx % total
	return m.hashMap[m.keys[rem]]
}
