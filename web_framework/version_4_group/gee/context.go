// gee/context.go

package gee

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Context struct {
    // origin objects
    Req    *http.Request
    Writer http.ResponseWriter
    // request info
    Path   string
    Method string
    Params map[string]string // 增加参数
    // response info
    StatusCode int
}

func NewContext(writer http.ResponseWriter, req *http.Request) *Context {
    return &Context{
        Req:    req,
        Writer: writer,
        Path:   req.URL.Path,
        Method: req.Method,
    }
}

func (c *Context) PostForm(key string) string {
    return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
    return c.Req.URL.Query().Get(key)
}

// Param 从请求地址中获取参数
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

func (c *Context) HTML(code int, html string) {
    c.SetHeader("Context-Type", "text/html")
    c.Status(code)
    if _, err := c.Writer.Write([]byte(html)); err != nil {
        http.Error(c.Writer, err.Error(), 500)
    }
}