package gee

import (
    "net/http"
    "strings"
)

type HandlerFunc func(ctx *Context)

type H map[string]interface{}

type Engine struct {
    *RouterGroup // 拥有 RouterGroup 的能力，可以隐形地调用 RouterGroup 的方法
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

    var middlewares []HandlerFunc
    for _, group := range engine.groups { // 遍历所有分组
        prefix := group.prefix + "/"
        if strings.HasPrefix(req.URL.Path, prefix) { // 查看当前请求，是否属于该分组
            middlewares = append(middlewares, group.middlewares...) // 属于该分组，将该分组的中间件保存下来
        }
    }

    context := NewContext(w, req)
    context.handlers = middlewares // 将保存下来的中间件，交给 context 依次执行
    engine.router.handle(context)
}

type RouterGroup struct {
    prefix      string
    middlewares []HandlerFunc // 储存该分组下的中间件
    engine      *Engine
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
    pattern = group.prefix + pattern
    group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
    group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
    group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) Use(handlerFunc ...HandlerFunc) { // 用户为该分组增加中间件
    group.middlewares = append(group.middlewares, handlerFunc...)
}