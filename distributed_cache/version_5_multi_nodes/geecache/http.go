package geecache

import (
	"fmt"
	"geecache/geecache/consistenthash"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/" // 请求路径应该是 "/<basepath>/<groupname>/<key>"
	defaultReplicas = 3
)

// HTTPPool :包含服务端和客户端，服务端提供获取本机数据的 HTTP 服务，客户端提供访问其他节点的方法
type HTTPPool struct {
	self     string // 当前节点的 IP/端口
	basePath string

	// 增加能力：设置和获取远程节点的能力
	mu          sync.Mutex
	peers       *consistenthash.Map    // 一致性哈希
	httpGetters map[string]*httpGetter // 记录每个远程节点的 httpGetter，httpGetter 包含了 baseURL
}

func NewHTTPPool(self string) *HTTPPool { // 为什么要设置这两个字段
	return &HTTPPool{
		self:     self,            // 本机的IP/端口
		basePath: defaultBasePath, // 请求前缀，便于过滤请求
	}
}

func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("Error path: " + r.URL.Path)
	}

	h.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(defaultBasePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(view.ByteSlice())
}

func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

// Set 设置远程节点的能力：实例化一个一致性哈希，并向哈希环中添加节点。
func (h *HTTPPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.peers = consistenthash.New(defaultReplicas, nil) // 初始化哈希环
	h.peers.Add(peers...)                              // 将节点加到哈希环上

	// 保存 key 和对应的 httpGetter 到 map 字段 httpGetters 中
	h.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetters[peer] = &httpGetter{
			baseURL: peer + h.basePath,
		}
	}
}

// PickPeer 获取远程节点的能力：包装了一致性哈希获取真实节点的方法 consistenthash.Map.Get
func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 从哈希环上获取缓存值属于那个节点，如果不是空且不是自身，则返回对应的节点。
	if peer := h.peers.Get(key); peer != "" && peer != h.self {
		h.Log("Pick peer %s", peer)
		return h.httpGetters[peer], true
	}
	return nil, false
}

// httpGetter 客户端：获取数据，实现 PeerGetter 接口
type httpGetter struct {
	baseURL string
}

// Get 步骤：
// 1.将 baseURL、group、key 拼接为远程节点缓存值的访问地址 URL
// 2.访问 URL 获取缓存值返回
func (g *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", g.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	/* url.QueryEscape 是 URL 转义函数，
	   比如 http://123.com/123?image=http://images.com/cat.png
	   需要转义为 http://123.com/123?image=http%3A%2F%2Fimages.com%2Fcat.png
	*/

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil) // 保证 httpGetter 实现了 PeerGetter。没实现在编译时报错。
