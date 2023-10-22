package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"version_6_template/gee"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.New()

	//r.Use(func(ctx *gee.Context) {
	//    now := time.Now()
	//    ctx.Next()
	//    fmt.Println(time.Since(now))
	//})

	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})

	stu1 := &student{Name: "Aim", Age: 20}
	stu2 := &student{Name: "Tao", Age: 22}
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "array.tmpl", gee.H{
			"title":        "gee",
			"studentArray": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
		})
	})

	r.Run(":9999")
}
