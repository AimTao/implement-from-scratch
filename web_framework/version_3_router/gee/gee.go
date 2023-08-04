package gee

import (
	"net/http"
)

type HandlerFunc func(ctx *Context)

type H map[string]interface{}

type Engine struct {
	router *router
}

func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.router.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.router.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine) // 传入 Handler 接口类型，只要实现了 ServeHTTP(ResponseWriter, *Request) 就实现了 Handler 接口。
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	context := NewContext(w, req)
	engine.router.handle(context)
}
