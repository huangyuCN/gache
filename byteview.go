package gache

//ByteView 只有一个成员变量b，它会存储真实的缓存值。
//选择byte类型是为了能够支持任意的数据类型的存储，例如字符串、图片等
type ByteView struct {
	b []byte
}

//Len 返回缓存数据的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个缓存数据切片的副本。b是只读的，防止缓存值被外部程序修改。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

//String 将缓存数据转换成string类型后返回。
func (v ByteView) String() string {
	return string(v.b)
}
