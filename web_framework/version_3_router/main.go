package main

import (
	"geeweb/gee"
	"net/http"
)

func main() {
	r := gee.New()

	r.GET("/hello", func(c *gee.Context) {
		// /hello?name=aimtao
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *gee.Context) {
		// /hello/aimtao
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gee.Context) {
		// /assets/pic/aimtao/1.jpg
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}

/*
curl http://127.0.0.1:9999/hello?name=aimtao
hello aimtao, you're at /hello

curl "http://127.0.0.1:9999/hello/aimtao?name=aimtao"
hello aimtao, you're at /hello/aimtao

curl http://127.0.0.1:9999/assets/pic/aimtao/1.jpg
{"filepath":"pic/aimtao/1.jpg"}

*/
