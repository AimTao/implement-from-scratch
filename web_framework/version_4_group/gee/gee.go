package gee

import (
    "net/http"
)

type HandlerFunc func(ctx *Context)

type H map[string]interface{}

type Engine struct {
    *RouterGroup
    router       *router
    groups       []*RouterGroup // 所有分组路径
}

func New() *Engine {
    engine := &Engine{router: newRouter()}
    engine.RouterGroup = &RouterGroup{engine: engine}
    engine.groups = []*RouterGroup{engine.RouterGroup}
    return engine
}

func (engine *Engine) Run(addr string) error {
    return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    context := NewContext(w, req)
    engine.router.handle(context)
}

type RouterGroup struct {
    prefix     string  // 该分组的完整前缀，例如 user 分组保存的是 "/v1/user/"，从最顶层的 group 到当前的 group
    middleware []HandlerFunc
    engine     *Engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
    newGroup := &RouterGroup{
        prefix: group.prefix + prefix,
        engine: group.engine,
    }
    group.engine.groups = append(group.engine.groups, newGroup)
    return newGroup
}

func (group *RouterGroup) addRoute(method string, pattern string, handler HandlerFunc) {
    pattern = group.prefix + pattern  // 拼接路径，比如上面例子中的 "/v1/user" 和 "/hello"
    group.engine.router.addRoute(method, pattern, handler)  // 间接使用 Engine 增加路由的能力
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
    group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
    group.addRoute("POST", pattern, handler)
}