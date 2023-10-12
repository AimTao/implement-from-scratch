package gee

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Context struct { // 暂且保存常用的参数
    // origin objects
    Req    *http.Request
    Writer http.ResponseWriter
    // request info
    Path   string
    Method string
    Params map[string]string
    // response info
    StatusCode int
    // middleware
    handlers []HandlerFunc
    index    int
    // engine pointer
    engine *Engine
}

func NewContext(writer http.ResponseWriter, req *http.Request) *Context {
    return &Context{
        Req:    req,
        Writer: writer,
        Path:   req.URL.Path,
        Method: req.Method,

        index: -1,
    }
}

func (c *Context) Next() {
    c.index++
    s := len(c.handlers)
    for ; c.index < s; c.index++ {
        c.handlers[c.index](c)
    }
}

func (c *Context) PostForm(key string) string {
    return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
    return c.Req.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
    return c.Params[key]
}

func (c *Context) SetHeader(key string, value string) {
    c.Writer.Header().Set(key, value)
}

func (c *Context) Status(code int) {
    c.StatusCode = code
    c.Writer.WriteHeader(code)
}

func (c *Context) String(code int, format string, values ...interface{}) {
    c.SetHeader("Content-Type", "text/plain")
    c.Status(code)
    if _, err := c.Writer.Write([]byte(fmt.Sprintf(format, values...))); err != nil {
        http.Error(c.Writer, err.Error(), 500)
    }
}

func (c *Context) JSON(code int, obj interface{}) {
    c.SetHeader("Context-Type", "application/json")
    c.Status(code)
    encoder := json.NewEncoder(c.Writer)
    if err := encoder.Encode(obj); err != nil {
        http.Error(c.Writer, err.Error(), 500)
    }
}

func (c *Context) Data(code int, date []byte) {
    c.Status(code)
    if _, err := c.Writer.Write(date); err != nil {
        http.Error(c.Writer, err.Error(), 500)
    }
}

func (c *Context) Fail(code int, err string) {
    c.index = len(c.handlers)
    c.JSON(code, H{"message": err})
}

func (c *Context) HTML(code int, name string, data interface{}) {
    c.SetHeader("Context-Type", "text/html")
    c.Status(code)

    // context 需要保存 Engine 指针，以便可以访问 htmlTemplates
    err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data)
    if err != nil {
        c.Fail(500, err.Error())
    }
}
