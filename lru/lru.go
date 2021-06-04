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
	maxBytes  int64                         //允许使用的最大内存
	nowBytes  int64                         //当前已使用的内存
	ll        *list.List                    //双向链表（标准库）
	cache     map[string]*list.Element      //字典，键是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) //当缓存数据被清理时回调，可以为nil
}

//entry 是双向链表的数据类型。在链表中任然保存key的好处在于，淘汰队首节点时，需要用key从字典中删除映射
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int //返回该值占用的内存大小
}

//New 是Cache的构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//Get 查询一个缓存记录，并将结果移动到队尾（双向链表作为队列，队首队尾是相对的，这里约定front为队尾）
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//RemoveOldest 删除过期缓存
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() //取到队首节点，从链表中删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                  //从字典中c.cache删除该节点的映射关系
		c.nowBytes -= int64(len(kv.key)) + int64(kv.value.Len()) //当前占用的内存
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) //回调
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { //更新已有缓存，并移动到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nowBytes += int64(value.Len()) + int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nowBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nowBytes {
		c.RemoveOldest()
	}
}

//Len 返回链表的长度（缓存个数）
func (c *Cache) Len() int {
	return c.ll.Len()
}
