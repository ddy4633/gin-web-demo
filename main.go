package main

import (
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"gin-web-demo/routes"
	"runtime"
)

func init() {
	//调优
	runtime.GOMAXPROCS(runtime.NumCPU())
	//初始化配置文件
	conf.InitConf()
	//初始化redis
	dao.NewRedis()
	//事件监听处理
	go saltstack.Event()
}

func main() {
	//初始化路由
	r := routes.SetupRouter()
	//获取当前的IP
	//ip := tools.GetHostIP()

	//启动服务
	if err := r.Run(":9090"); err != nil {
		fmt.Println(err)
		return
	}
}
