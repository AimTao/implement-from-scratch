// gee/router.go

package gee

import "fmt"

type router struct {
	roots    map[string]*node       // 不同请求方式的 Trie 树根节点
	handlers map[string]HandlerFunc // 储存路由对应的请求处理函数
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// 增加路由映射
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	// 添加请求方法，例如 GET、POST
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{children: make(map[string]*node)}
	}

	// 路由地址插入前缀树
	r.roots[method].insert(pattern)

	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) getRouter(method string, pattern string) (*node, map[string]string) {
	// 如果不存在该请求方法，直接退出
	if _, ok := r.roots[method]; !ok {
		return nil, nil
	}

	return r.roots[method].search(pattern)
}

func (r *router) handle(c *Context) {
	node, params := r.getRouter(c.Method, c.Path)
	if node != nil {
		c.Params = params
		key := c.Method + "-" + node.path
		r.handlers[key](c)
	} else {
		fmt.Fprintf(c.Writer, "404 NOT FOUND: %s\n", c.Req.URL)
	}
}
