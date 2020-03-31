package main

import (
	"fmt"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"gin-web-demo/routes"
	"github.com/gin-gonic/gin"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	//初始化redis
	dao.NewRedis()
	route := gin.Default()
	//初始化日志信息
	//route.Use(logs.LoggerToFile())
	//定义数据的增删改查操作
	group_inter := route.Group("/data")
	group_inter.POST("/add")
	group_inter.POST("/delete")
	group_inter.GET("/query/:id", routes.GetQueryJobInfo)
	//静态资源加载
	route.Static("/static", "static")
	//加载模板
	route.LoadHTMLGlob("views/*")
	//定义常规路由
	//greoup_v1 := route.Group("v1")
	route.POST("/monitor", routes.AlterManagerWebHookHandler)
	route.GET("/token", routes.GetToken)
	route.POST("/config", routes.PostJobhandler)
	route.GET("/jobs", routes.GetJobInfo)
	//事件监听处理
	go saltstack.Event()
	//获取当前的IP
	//ip := tools.GetHostIP()
	//启动服务
	if err := route.Run(":9090"); err != nil {
		fmt.Println(err)
		return
	}
}
