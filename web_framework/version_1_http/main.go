package main

import (
	"fmt"
	"geeweb/gee"
	"net/http"
)

func main() {
	engine := gee.New()
	engine.GET("/", func(w http.ResponseWriter, req *http.Request) { // 注册静态路由
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	engine.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	err := engine.Run(":9999")
	if err != nil {
		panic(err)
	}
}
