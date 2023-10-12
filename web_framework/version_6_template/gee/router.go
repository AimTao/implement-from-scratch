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
    // 添加请求方法，例如 GET、POST
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
        c.handlers = append(c.handlers, r.handlers[key])
    } else {
        fmt.Fprintf(c.Writer, "404 NOT FOUND: %s\n", c.Req.URL)
    }
    c.Next()
}
