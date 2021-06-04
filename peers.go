package gache

import pb "gache/gachepb"

// PeerPicker 的 PickPeer 方法用于根据传入的key选择相应节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 的 Get 方法用于从对应group查找缓存。PeerGetter对应HTTP客户端
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
