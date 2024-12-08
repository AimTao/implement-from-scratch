package main

import (
	"geeweb/gee"
	"net/http"
)

func main() {
	engine := gee.New()
	v1 := engine.Group("/v1")
	v1.GET("/", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "<h1>hello Gee</h1>")
	})

	v1.GET("/hello", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "hello %s", ctx.Query("name"))
	})

	v1.POST("/login", func(ctx *gee.Context) {
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
curl http://127.0.0.1:9999/v1
<h1>hello Gee</h1>

curl http://127.0.0.1:9999/v1/hello?name=aimtao
hello aimtao

curl http://127.0.0.1:9999/v1/login -X POST -d 'username=aimtao&password=123'
{"username":"aimtao", "password":"123"}
*/
