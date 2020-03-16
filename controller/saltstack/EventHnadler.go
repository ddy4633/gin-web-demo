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

func sched(data *conf.EventHand) {
	info := reddao.GetDate("token")
	//构造函数
	Job := &conf.JobRunner{
		Client: "local_async",
		Tgt:    data.Address,
		Fun:    "cmd.run",
		Arg:    "ping -c 2 baidu.com",
	}
	resultid := a.PostModulJob(info, Job)
	conf.Chan2 <- resultid
}

func handl(info *conf.JobReturn) {
	count := info.Count
	count++
	token := reddao.GetDate("token")
	if count == 20 {
		fmt.Println(info.Return[0].Minions, "+", info.Return[0].Jid, "无法获取到JOb信息")
		return
	}
	data := a.QueryJob(info.Return[0].Jid, token)
	fmt.Println("!!!!!!!!时间处理的data=", data)
	key := info.Return[0].Minions[0]
	//fmt.Println("key=",key,"return=",data.Info[0].Result[key])
	if _, ok := data.Info[0].Result[key]; ok {
		fmt.Println("序列化前:", data)
		//序列化
		data, err := json.Marshal(&info)
		if !conf.CheckERR(err, "EventHandle JSON Marshal is Failed") {
			return
		}
		//写入redis数据库
		if err := reddao.InsertDate(info.Return[0].Jid, string(data)); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("插入数据库成功=", info.Return[0].Jid)
	} else {
		info.Count = count
		fmt.Println("没有获取到", info.Return[0].Minions, info.Count, "ID=", info.Return[0].Jid)
		conf.Chan2 <- info
	}
}
