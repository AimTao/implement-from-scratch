package main

import (
	"net/http"
	"version_2_context/gee"
)

func main() {
	engine := gee.New()
	engine.GET("/", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "<h1>hello Gee</h1>")
	})

	engine.GET("/hello", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "hello %s", ctx.Query("name"))
	})

	engine.POST("/login", func(ctx *gee.Context) {
		ctx.JSON(http.StatusOK, gee.H{
			"username": ctx.PostForm("username"),
			"password": ctx.PostForm("password"),
		})
	})

	err := engine.Run(":9999")
	if err != nil {
		panic(err)
	}
}

/*
curl http://127.0.0.1:9999/
<h1>hello Gee</h1>

curl http://127.0.0.1:9999/hello?name=aimtao
hello aimtao

curl http://127.0.0.1:9999/hello -X POST -d 'username=aimtao&password=123'
{"username":"aimtao", "password":"123"}
*/
