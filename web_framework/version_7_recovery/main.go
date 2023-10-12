package main

import "version_6_template/gee"

func main() {
    r := gee.New()

    r.Use(gee.Recovery())

    r.GET("/", func(ctx *gee.Context) {
        panic("err")
    })

    r.Run(":9999")
}
