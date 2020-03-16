package saltstack

import (
	"fmt"
	"gin-web-demo/conf"
)

func Event() {
	for  {
		select  {
		case data := <-conf.Chan1:
			go sched(data)
		}
	}
}

func EventInfo() {
	for {
		select {
		case info := <-conf.Chan2:
			go handl(info)
		}
	}
}

func sched(data *conf.EventHand) {
	a := SaltController{}
	info := a.GetToken()
	conf.Token = info.Return[0].Token
	//构造函数
	Job := &conf.JobRunner{
		Client:    "local_async",
		Tgt:       data.Address,
		Fun:       "cmd.run",
		Arg:       "ping -c 2 baidu.com",
	}
	resultid  := a.PostModulJob(conf.Token,Job)
	conf.Chan2 <- resultid
}

func handl(info *conf.JobReturn) {
	a := SaltController{}
	count := info.Count
	count += 1
	if count == 20 {
		fmt.Println(info.Return[0].Minions,"+",info.Return[0].Jid,"无法获取到JOb信息")
		return
	}
	data := a.QueryJob(info.Return[0].Jid,conf.Token)
	key := info.Return[0].Minions[0]
	if _,ok := data.Info[0].Result[key]; ok {
		fmt.Println("当前次数:",info.Count,"\n",data.Info[0].StartTime,"\n",data.Info[0].Result)
	}else {
		info.Count = count
		conf.Chan2 <- info
	}
}