package routes

import (
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
)

var redao = dao.RedisHandle{}

//定义Webhook
func AlterManagerWebHookHandler(c *gin.Context) {
	var (
		info conf.HookMessageInfo
	)
	//转换字节流
	byte, _ := ioutil.ReadAll(c.Request.Body)
	//转换成json对象
	if json.Unmarshal(byte, &info) != nil {
		fmt.Println("json error")
	}
	//定义接受对象
	var ent *conf.EventHand
	//赋值对象
	ent = &conf.EventHand{
		EventName: info.Alerts[0].Labels["lables"],
		HostName:  info.Alerts[0].Labels["hostname"],
		Address:   strings.Split(info.Alerts[0].Labels["instance"], ":")[0],
		Event:     info.Alerts[0].Annotations["description"],
		Status:    1,
	}
	//传递给channel调用
	conf.Chan1 <- ent
	c.Writer.WriteString("ok")
}

//测试Token
func GetToken(c *gin.Context) {
	//var data conf.Returninfo
	a := saltstack.SaltController{}
	data := a.GetToken()
	//设置Token的过期时间
	err := redao.InsertTTLData("token", data.Return[0].Token, "EX", "86400")
	if !conf.CheckERR(err, "redisDAO SET Token is Failed") {
		c.Writer.WriteString("写入Token失败")
	}
	c.JSON(200, gin.H{"toekn": data.Return[0].Token})
}

//执行命令
func PostJobhandler(c *gin.Context) {
	//接受post任务的参数
	cli := c.PostForm("client")
	tgt := c.PostForm("tgt")
	expr := c.PostForm("expr_form")
	fun := c.PostForm("fun")
	arg := c.PostForm("arg")

	//构造函数
	Job := &conf.JobRunner{
		Client:    cli,
		Tgt:       tgt,
		Expr_form: expr,
		Fun:       fun,
		Arg:       arg,
	}
	//将执行的信息序列化存储到后端的redis中
	data, err := json.Marshal(Job)
	if !conf.CheckERR(err, "[PostJobHandler] json Marshal is Failed") {
		c.JSON(401, gin.H{
			"status": 1,
			"info":   err,
		})
		return
	}
	//插入数据库
	err = redao.InsertDate("Config", string(data))
	if !conf.CheckERR(err, "[PostJobHandler] Insert Redis is Failed") {
		c.JSON(401, gin.H{
			"status": 1,
			"info":   err,
		})
		return
	}
	c.JSON(200, gin.H{
		"status": 0,
	})
}

//获取执行后的任务信息(以Josn回写)
func GetJobInfo(c *gin.Context) {
	id := c.Request.FormValue("id")
	fmt.Println(id)
	data := redao.GetDate(id)
	fmt.Println(data)
	c.Writer.WriteHeader(200)
	c.Writer.WriteString(data)
}

//获取指定数量的Job任务数
func GetJobListPage(c *gin.Context) {
	//页面
	//page := c.Request.FormValue("page")

}
