package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func Logger() gin.HandlerFunc{
	return func(c *gin.Context){
		t:=time.Now()
		c.Set("example","hi!这是一个中间件数据")
		c.Next()

		latency:=time.Since(t)
		log.Print(latency)

		status:=c.Writer.Status()
		log.Println(status)
	}
}

func main(){
	r:=gin.New()
	r.Use(Logger())
	r.GET("/hi",func(c *gin.Context){
		example:=c.MustGet("example").(string)
		log.Println(example)
	})
	r.Run(":8080")
}
