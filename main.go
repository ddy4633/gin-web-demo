package main

import (
	"fmt"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	//初始化
	//a := make(chan conf.EventsWait)

	route := gin.Default()
	//定义路由
	route.POST("/monitor", routes.AlterManagerWebHookHandler)
	route.GET("/token", routes.GetToken)
	route.POST("/test", routes.PostJobhandler)
	//事件
	go saltstack.Event()
	//
	go saltstack.EventInfo()
	//启动服务
	if err := route.Run(":9091"); err != nil {
		fmt.Println(err)
		return
	}
}
