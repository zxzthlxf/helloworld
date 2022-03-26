package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main(){
	router:=gin.Default()
	router.MaxMultipartMemory=64<<20
	router.POST("/upload",func(c *gin.Context){
		file,_:=c.FormFile("file")
		log.Println(file.Filename)

		c.SaveUploadedFile(file,"./data/shirdon.jpg")
		c.String(http.StatusOK,fmt.Sprintf("'%s' uploaded!",file.Filename))
	})
	router.Run(":8086")
}
