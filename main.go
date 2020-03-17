package main

import (
	"fmt"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"gin-web-demo/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	//初始化redis
	dao.NewRedis()
	route := gin.Default()
	//定义路由
	route.POST("/monitor", routes.AlterManagerWebHookHandler)
	route.GET("/token", routes.GetToken)
	route.POST("/test", routes.PostJobhandler)
	route.GET("/jobs/", routes.GetJobInfo)
	//事件监听处理
	go saltstack.Event()
	//启动服务
	if err := route.Run(":9090"); err != nil {
		fmt.Println(err)
		return
	}
}
