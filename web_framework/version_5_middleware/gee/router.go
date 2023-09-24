package gee

import "fmt"

type router struct {
    roots    map[string]*node
    handlers map[string]HandlerFunc
}

func newRouter() *router {
    return &router{
        roots:    make(map[string]*node),
        handlers: make(map[string]HandlerFunc),
    }
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
    if _, ok := r.roots[method]; !ok {
        r.roots[method] = &node{children: make(map[string]*node)}
    }

    r.roots[method].insert(pattern)

    key := method + "-" + pattern
    r.handlers[key] = handler

    fmt.Println("key", key)
}

func (r *router) getRouter(method string, pattern string) (*node, map[string]string) {
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
        c.handlers = append(c.handlers, r.handlers[key]) // 将找到的 请求处理函数 加在 c.handlers 的后面
    } else {
        fmt.Fprintf(c.Writer, "404 NOT FOUND: %s\n", c.Req.URL)
    }
    c.Next() // 依次执行 c.handlers 的函数。
    // c.Next() 一定要放在 这里即使 404，也要执行中间件
}