package saltstack

import (
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/dao"
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
			go handl(info)
		}
	}
}

//临时指定要post的操作
func sched(data *conf.EventHand) {
	info := reddao.GetDate("token")
	//构造函数
	Job := &conf.JobRunner{
		Client: "local_async",
		Tgt:    data.Address,
		Fun:    "cmd.run",
		Arg:    "time ping -c 2 baidu.com",
	}
	resultid := a.PostModulJob(info, Job)
	conf.Chan2 <- resultid
}

//执行jobs事件的查询
func handl(info *conf.JobReturn) {
	//统计查询的次数
	count := info.Count
	count++
	//查询Token
	token := reddao.GetDate("token")
	//执行任务的ID号
	jid := info.Return[0].Jid
	//中断指令
	if count == 60 {
		reddao.SaddDate(info.Return[0].Jid)
		fmt.Println(info.Return[0].Minions, "+", jid, "无法获取到JOb信息")
		return
	}
	//查询任务情况
	data := a.QueryJob(jid, token)
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
		fmt.Println("插入数据库成功=", info.Return[0].Jid)
	} else {
		//info.Count = count
		//fmt.Println("没有获取到", info.Return[0].Minions, info.Count, "ID=", info.Return[0].Jid)
		conf.Chan2 <- info
	}
}
