package main

import (
    "fmt"
    "net/http"
    "time"
    "version_5_middleware/gee"
)

func main() {
    r := gee.New()

    v1 := r.Group("/v1")
    v1.Use(func(ctx *gee.Context) {
        now := time.Now()
        ctx.Next()
        fmt.Println(time.Since(now))
    })

    user := v1.Group("/user")
    user.GET("/hello/:name", func(ctx *gee.Context) {
        ctx.JSON(http.StatusOK, gee.H{
            "name": ctx.Param("name"),
        })
    })


    err := r.Run(":9999")
    if err != nil {
        panic(err)
    }

}

// curl "http://127.0.0.1:9999/v1/user/hello/aimtao"
// {"name":"aimtao"}