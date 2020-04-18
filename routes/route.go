package routes

import (
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"gin-web-demo/tools"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"
	"time"
)

var redao = dao.RedisHandle{}
var salt = saltstack.SaltController{}

//定义Webhook
func AlterManagerWebHookHandler(c *gin.Context) {
	var (
		test conf.Notification
	)
	err := c.BindJSON(&test)
	if !tools.CheckERR(err, "receive alterManager info json is Failed") {
		return
	}
	//test过程
	ent1 := &conf.EventHand{
		EventName: test.Alerts[0].Labels["lables"],
		HostName:  test.Alerts[0].Labels["hostname"],
		Address:   strings.Split(test.Alerts[0].Labels["instance"], ":")[0],
		Event:     test.Alerts[0].Annotations["description"],
		Status:    1,
	}
	//生成全局唯一ID
	id := tools.GetMd5String(tools.GetTimeNow())
	//构造初始对象
	Obj := &conf.AllMessage{ID: id, Notifications: test, Eventhand: ent1}
	//序列化
	data, err := json.Marshal(Obj)
	if !tools.CheckERR(err, "JSON marshal ALLObj is Failed") {
		return
	}
	//插入数据库
	err = dao.RedisHandle{}.InsertDate(id, string(data))
	if !tools.CheckERR(err, "Insert Redis dao.ALLObj is Failed") {
		return
	}
	//传递给channel调用
	conf.Chan1 <- Obj
	conf.WriteLog(fmt.Sprintf("%s[Return]新事件进入ID为=%s\n", time.Now().Format("2006-01-02 15:04:05"), id))
	c.Writer.WriteString("ok")
}

//测试Token
func GetToken(c *gin.Context) {
	//var data conf.Returninfo
	a := saltstack.SaltController{}
	data := a.GetToken()
	//设置Token的过期时间
	err := redao.InsertTTLData("token", data.Return[0].Token, "EX", "18000")
	if !tools.CheckERR(err, "redisDAO SET Token is Failed") {
		c.Writer.WriteString("写入Token失败")
	}
	c.JSON(200, gin.H{"toekn": data.Return[0].Token})
}

//执行命令
func PostJobhandler(c *gin.Context) {
	//接受post任务的参数
	var (
		job       conf.AddonJobRunner
		runner    conf.JobRunner
		paraMeter conf.ParaMeter
		err       error
		data      []byte
	)
	//指定结构体的映射
	err = c.ShouldBindBodyWith(&job, binding.JSON)
	if !tools.CheckERR(err, "[ROUTE PostJobhandler] Set Config BindJson is Failed ") {
		goto TARGET
	}
	err = c.ShouldBindBodyWith(&paraMeter, binding.JSON)
	if !tools.CheckERR(err, "[ROUTE PostJobhandler] Set Config BindJson is Failed ") {
		goto TARGET
	}
	err = c.ShouldBindBodyWith(&runner, binding.JSON)
	if !tools.CheckERR(err, "[ROUTE PostJobhandler] Set Config BindJson is Failed ") {
		goto TARGET
	}
	//对象变量赋值
	job.Job = runner
	job.Para = paraMeter
	//将执行的信息序列化存储到后端的redis中
	data, err = json.Marshal(job)
	if !tools.CheckERR(err, "[PostJobHandler] json Marshal is Failed") {
		c.JSON(401, gin.H{
			"status": 1,
			"info":   err,
		})
		return
	}
	//插入数据库
	err = redao.InsertDate("Config", string(data))
	if !tools.CheckERR(err, "[PostJobHandler] Insert Redis is Failed") {
		c.JSON(401, gin.H{
			"status": 1,
			"info":   err,
		})
		return
	}
	c.JSON(200, gin.H{
		"status": 0,
		"job":    job,
	})
	return
	//错误处理
TARGET:
	c.JSON(400, gin.H{
		"message": err.Error(),
		"code":    1,
		"hint":    "请检查config配置文件参数是否正确或是缺失!",
	})
	return
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

func GetQueryJobInfo(c *gin.Context) {
	//取出id数据
	id := c.Param("id")
	//查数据库
	data := dao.RedisHandle{}.GetDate(id)
	info := &conf.EndJob{}
	json.Unmarshal([]byte(data), info)
	//c.JSON(200,info)
	c.HTML(200, "Message.html", info)
}

//获取指定数量的Job任务数
func GetJobListPage(c *gin.Context) {
	//页面
	//page := c.Request.FormValue("page")

}

//心跳
func ResponPong(c *gin.Context) {
	c.JSON(200, gin.H{"Message": "pong"})
}

//测试
func Textfun(c *gin.Context) {
	job := conf.JobRunner{}
	err := c.BindJSON(&job)
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	c.JSON(200, job)
}

//salt-minion存活检测
func CheckActive(c *gin.Context) {
	adress := c.PostForm("address")
	status, err := salt.ActiveSalt(adress)
	if status {
		c.JSON(200, gin.H{"address": adress, "active": status, "message": err})
	} else {
		c.JSON(400, gin.H{"address": adress, "active": status, "message": err})
	}

}
