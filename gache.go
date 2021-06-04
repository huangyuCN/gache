package gache

import (
	"fmt"
	pb "gache/gachepb"
	"gache/singleflight"
	"log"
	"sync"
)

//Group 缓存的命名空间，每个Group有一个唯一的名称name。
type Group struct {
	name      string              //缓存空间名字
	getter    Getter              //缓存未命中时从数据源获取数据的回调
	mainCache cache               //真正的缓存信息
	peers     PeerPicker          //远程节点
	loader    *singleflight.Group //使用singleflight.Group保证同一个key的请求只会被执行一次
}

//Getter 是一个接口，用于在无法找到缓存数据的时候从其他数据源加载进缓存
type Getter interface {
	Get(key string) ([]byte, error)
}

//GetterFunc 是一个函数类型，并实现 Getter 接口的 Get 方法。
type GetterFunc func(key string) ([]byte, error)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) //全局变量，保存所有命名空间
)

//Get 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，\
//也能够传入实现了该接口的结构体作为参数。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//NewGroup 创建Group对象的构造函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

//GetGroup 用来查找特定的命名空间
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

//Get 查找缓存，如果找到直接返回，如果没找到则通过回调函数从数据源查找
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Gache Cache] hit")
		return v, nil
	}
	return g.load(key)
}

//RegisterPeers 实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

//使用 PickPeer() 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。
//若是本机节点或失败，则回退到 getLocally()
func (g *Group) load(key string) (value ByteView, err error) {
	viewInterface, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Gcache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewInterface.(ByteView), err
	}
	return
}

//通过调用回调函数，从源查找数据，并添加到缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

//getFromPeer 从远程获取
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

//将数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
