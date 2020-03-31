package saltstack

import (
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	dd "gin-web-demo/controller/dingding"
	"gin-web-demo/dao"
	"strings"
)

var (
	reddao = dao.RedisHandle{}
	a      = SaltController{}
)

//事件处理流
func Event() {
	for {
		select {
		case data := <-conf.Chan1:
			go sched(data)
		case info := <-conf.Chan2:
			if info.JobReceipt.Return == nil {
				continue
			} else {
				go handl(info)
			}
		}
	}
}

//临时指定要post的操作
func sched(data *conf.AllMessage) {
	var (
		job      *conf.AddonJobRunner
		info, ac string
	)
	//获取Token信息
	if info = reddao.GetDate("token"); info == "" {
		info = a.GetToken().Return[0].Token
		err := reddao.InsertTTLData("token", info, "EX", "86400")
		if !conf.CheckERR(err, "Inserter Token Failed") {
			return
		}
	}
	//获取指定的参数信息
	if ac = reddao.GetDate("Config"); ac == "" {
		fmt.Println("请设置好Config信息")
		return
	}
	//反序列化得到变量
	err := json.Unmarshal([]byte(ac), &job)
	if !conf.CheckERR(err, "") {
		return
	}
	//处理事件过滤
	if !filtstring(data.Eventhand, job.Para) {
		return
	}
	//判断事件是否需要处理
	if job.Switch == 1 {
		fmt.Printf("[Info]开关已经关闭,当前事件为=%s\n", data)
		return
	}
	job.Job.Tgt = data.Eventhand.Address
	//赋值对象
	data.AddonRunners = job
	/*测试的时候使用
	  //构造函数
	  Job := &conf.JobRunner{
	          Client: "local_async",
	          Tgt:    data.Address,
	          Fun:    "cmd.run",
	          Arg:    "time ping -c 2 baidu.com",
	  }
	*/
	//进行Post请求取回事物执行ID
	resultid := a.PostModulJob(info, &job.Job)
	//构造对象
	data.JobReceipt = resultid
	//fmt.Println(data.Event);
	conf.Chan2 <- data
}

//执行jobs事件的查询
func handl(info *conf.AllMessage) {
	//统计查询的次数
	count := info.JobReceipt.Count
	count++
	//查询Token
	token := reddao.GetDate("token")
	//执行任务的ID号
	jid := info.JobReceipt.Return[0].Jid
	//中断指令
	if count == info.AddonRunners.Count {
		reddao.SaddDate(info.JobReceipt.Return[0].Jid)
		fmt.Println(info.JobReceipt.Return[0].Minions, "+", jid, "无法获取到JOb信息")
		return
	}
	//查询任务情况
	data := a.QueryJob(jid, token)
	//排除空数组行为
	if data.Info[0].Minions == nil {
		return
	}
	//获取目标主机的IP
	//key := info.Return[0].Minions[0]
	//判断是否取值成功,失败则重新进入队列等待再次的处理
	if data.Info[0].Result[data.Info[0].Minions[0]].Success {
		//构造写入redis的数据信息
		endjob := conf.SetData(data)
		a, err := json.Marshal(endjob)
		if !conf.CheckERR(err, "构造写入redis的数据信息Json Manshal is Failed") {
			return
		}
		//写入redis数据库(data.Info[0].Result[key].Return)
		if err := reddao.InsertDate(jid, string(a)); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("插入数据库成功=", info.JobReceipt.Return[0].Jid)
		//传递的钉钉构造函数
		markdown := conf.SetDD(info, endjob)
		err = dd.Postcontent(markdown)
		if !conf.CheckERR(err, "PostDingding is Failed") {
			return
		}
	} else {
		//info.JobReceipt.Count = count
		//fmt.Println("没有获取到", info.JobReceipt.Return[0].Minions, info.JobReceipt.Count, "ID=", info.JobReceipt.Return[0].Jid)
		conf.Chan2 <- info
	}
}

//过滤处理的事件
func filtstring(data *conf.EventHand, para conf.ParaMeter) bool {
	//取出事件
	event := data.Event
	//取出主机名称
	hostname := data.HostName
	//切分过滤的数据
	ev := strings.Split(para.ParaEvent, ",")
	//过滤所有的字段是否匹配
	for _, a := range ev {
		if a != "" && strings.Contains(event, a) {
			return false
		}
	}
	//Hostname判断
	evhost := strings.Split(para.ParaHost, ",")
	for _, b := range evhost {
		if b != "" && strings.Contains(hostname, b) {
			return false
		}
	}
	return true
}
