package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil { // 捕获 panic error
				log.Printf("%s\n\n", trace(err))                                // 解析 error
				c.Fail(http.StatusInternalServerError, "Internal Server Error") // 给用户返回 500 错误
			}
		}()
		c.Next() // 为了保证 defer 的正确执行顺序，务必需要写 c.Next
	}
}

func trace(error interface{}) string {
	message := fmt.Sprintf("%s", error)
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
