package gee

import "fmt"

type router struct {
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
	}
}

// 增加路由映射
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// 封装统一的调用请求处理函数的方法，供 ServeHTTP 调用。
func (r *router) handle(c *Context) { // 注意参数已更改为 context。
	key := c.Req.Method + "-" + c.Req.URL.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		fmt.Fprintf(c.Writer, "404 NOT FOUND: %s\n", c.Req.URL)
	}
}
