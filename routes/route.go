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
	"time"
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
		Event:     info.Alerts[0].Annotations["annotations"],
		Status:    1,
	}
	//传递给channel调用
	conf.Chan1 <- ent
	c.Writer.WriteString("ok")
}

//测试去Token
func GetToken(c *gin.Context) {
	//var data conf.Returninfo
	a := saltstack.SaltController{}
	data := a.GetToken()
	//设置Token的过期时间
	err := redao.InsertTTLData("token", data.Return[0].Token, "EX", "86400")
	if !conf.CheckERR(err, "redisDAO SET Token is Failed") {
		c.Writer.WriteString("写入Token失败")
	}
	c.Writer.WriteHeader(200)
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

	a := saltstack.SaltController{}
	//获取token
	token := redao.GetDate("token")
	reinfo := a.PostModulJob(token, Job)
	time.Sleep(5 * time.Second)
	a.QueryJob(reinfo.Return[0].Jid, token)
}

//获取执行后的任务信息
func GetJobInfo(c *gin.Context) {
	id := c.Request.FormValue("id")
	data := redao.GetDate(id)
	c.Writer.WriteHeader(200)
	c.Writer.WriteString(data)
}
