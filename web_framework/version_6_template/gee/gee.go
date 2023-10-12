package gee

import (
    "html/template"
    "net/http"
    "path"
    "strings"
)

type HandlerFunc func(ctx *Context)

type H map[string]interface{}

type Engine struct {
    *RouterGroup
    router *router
    groups []*RouterGroup

    // for html render
    htmlTemplates *template.Template // 将所有模板加载进内存
    funcMap       template.FuncMap   // 模板的渲染函数(可自定义)
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
    for _, group := range engine.groups {
        prefix := group.prefix + "/"
        if strings.HasPrefix(req.URL.Path, prefix) {
            middlewares = append(middlewares, group.middlewares...)
        }
    }

    context := NewContext(w, req)
    context.handlers = middlewares
    context.engine = engine
    engine.router.handle(context)
}

// SetFuncMap 设置渲染函数，可以在模板中指定，某个数据使用某个渲染函数
// 传入的 template.FuncMap，一个 map，保存了渲染函数对应的名称，在模板中使用名称即可指定渲染函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
    engine.funcMap = funcMap
}

// LoadHTMLGlob 指定模板的路径，将模板加载到内存中
func (engine *Engine) LoadHTMLGlob(pattern string) {
    engine.htmlTemplates = template.New("")
    engine.htmlTemplates.Funcs(engine.funcMap)
    engine.htmlTemplates = template.Must(engine.htmlTemplates.ParseGlob(pattern))
}

type RouterGroup struct {
    prefix      string
    middlewares []HandlerFunc
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

func (group *RouterGroup) Use(handlerFunc ...HandlerFunc) {
    group.middlewares = append(group.middlewares, handlerFunc...)
}

//Static 例如 r.Static("/assets", "./static")
func (group *RouterGroup) Static(relativePath string, root string) {
    urlPattern := path.Join(relativePath, "/*filepath")                // 1.拼接路径，例如得到 "/assets/*filepath"
    handler := group.createStaticHandler(relativePath, http.Dir(root)) // 2.得到路由处理函数 handler
    group.GET(urlPattern, handler)                                     // 3.添加路由映射，例如将路由地址 "/assets/*filepath" 和处理函数 handler 相绑定。
}

// createStaticHandler 例如获取路由地址 "/assets/*filepath" 的处理函数
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
    absolutePath := path.Join(group.prefix, relativePath) // 拼接上路由分组的前缀, 例如 "/user" + "/assets", 即得到完整前缀，不包含实际文件路径

    /* fileServer 是一个handler 接口类型变量，它的功能是：
       1.将请求的 URL 中的前缀 absolutePath 去掉得到文件路径
       2.将文件路径交给 http.FileServer(fs) 这个handler 来打开。
    */
    fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

    return func(ctx *Context) {
        file := ctx.Param("filepath")
        if _, err := fs.Open(file); err != nil {
            ctx.Status(http.StatusNotFound)
            return
        }
        fileServer.ServeHTTP(ctx.Writer, ctx.Req) // 这里调用 fileServer.ServeHTTP，实际在调用 fileServer 函数本身。
    }
}
