package main

import (
	"flag"
	"fmt"
	"gache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gache.Group {
	return gache.NewGroup("scores", 2<<10, gache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

//startCacheServer 用来启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 gee 中，启动 HTTP 服务
func startCacheServer(addr string, addrs []string, gee *gache.Group) {
	peers := gache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("gache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))

}

//startAPIServer 用来启动一个 API 服务（端口 9999），与用户进行交互
func startAPIServer(apiAddr string, gee *gache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// 函数需要命令行传入 port 和 api 2 个参数，用来在指定端口启动 HTTP 服务
// 为了方便，我们将启动的命令封装为一个 shell 脚本
func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Gache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
