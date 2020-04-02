package routes

import (
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/controller/saltstack"
	"gin-web-demo/dao"
	"gin-web-demo/tools"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

var redao = dao.RedisHandle{}

//定义Webhook
func AlterManagerWebHookHandler(c *gin.Context) {
	var (
		//info conf.HookMessageInfo
		test conf.Notification
	)
	//转换字节流
	//byte, _ := ioutil.ReadAll(c.Request.Body)
	//转换成json对象
	//if json.Unmarshal(byte, &info) != nil {
	//	fmt.Println("json error")
	//}
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
	cli := c.PostForm("client")
	tgt := c.PostForm("tgt")
	expr := c.PostForm("expr_form")
	fun := c.PostForm("fun")
	arg := c.PostForm("arg")
	parahost := c.PostForm("parahost")
	paraevent := c.PostForm("paraevent")
	switchs := c.PostForm("switch")
	count := c.PostForm("count")
	aint, _ := strconv.Atoi(switchs)
	cint, _ := strconv.Atoi(count)
	//构造函数
	info := conf.JobRunner{
		Client:    cli,
		Tgt:       tgt,
		Expr_form: expr,
		Fun:       fun,
		Arg:       arg,
	}
	para := conf.ParaMeter{
		ParaHost:  parahost,
		ParaEvent: paraevent,
	}
	Job := &conf.AddonJobRunner{
		Job:    info,
		Para:   para,
		Switch: aint,
		Count:  cint,
	}
	//将执行的信息序列化存储到后端的redis中
	data, err := json.Marshal(Job)
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
