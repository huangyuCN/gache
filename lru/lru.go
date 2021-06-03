/**
Package lru LRU(Least Recently Used)
最近最少使用，相对于仅考虑时间因素的 FIFO(First In First Out)和仅考虑访问频率的LFU(Least Frequently Used)，
LRU 算法可以认为是相对平衡的一种淘汰算法。LRU认为，如果数据最近被访问过，那么将来被访问的概率也会更高。
LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，
淘汰该条记录即可。
 */

package lru

import "container/list"

// Cache 是一个LRU缓存，它是并发不安全的
type Cache struct {
	maxBytes int64
	nbytes int64
	ll *list.List
	cache map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key string
	value Value
}

type Value interface {
	Len() int
}