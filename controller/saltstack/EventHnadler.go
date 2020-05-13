package saltstack

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-web-demo/conf"
	dd "gin-web-demo/controller/dingding"
	"gin-web-demo/dao"
	"gin-web-demo/tools"
	"log"
	"reflect"
	"strings"
	"time"
)

var (
	reddao = dao.RedisHandle{}
	salt   = SaltController{}
	active = "minion存活检查有问题"
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
		case dmsg := <-conf.ChanDD:
			dd.Postcontent(dmsg)
		case result := <-conf.ChanJobs:
			ApiSalt(result)
		}
	}
}

//临时指定要post的操作
func sched(data *conf.AllMessage) {
	var (
		job      *conf.AddonJobRunner
		info, ac string
	)
	//判断是否是重试
	if data.AddonRunners.RetrySwitch == 1 {
		return
	}
	//记录时间
	data.TimeTotles.BeginTime = time.Now()
	//延迟推出后删除队列
	defer reddao.ZremDate("Eventlist", data.Eventhand.Address)
	//获取Token信息
	info, err := salt.Check()
	tools.CheckERR(err, "取回token失败")
	//获取auth Token
	if auth := reddao.GetDate("AuthToken"); auth == "" {
		err := salt.GetCMDBAUTH()
		if err != nil {
			//conf.WriteLog(fmt.Sprintf("[error]无法获取到token,%s\n", err))
			log.Printf("无法获取到token,%s\n", err)
			return
		}
	}
	//获取指定的参数信息
	if ac = reddao.GetDate("Config"); ac == "" {
		//conf.WriteLog(fmt.Sprintf("[info]Config配置信息没有设置\n"))
		log.Printf("Config配置信息没有设置\n")
		return
	}
	//反序列化得到变量
	err = json.Unmarshal([]byte(ac), &job)
	if !tools.CheckERR(err, "") {
		return
	}
	//处理事件过滤
	if !filtstring(data.Eventhand, job.Para) {
		return
	}
	//判断事件是否需要处理
	if job.Switch == 1 {
		//conf.WriteLog(fmt.Sprintf("%s[info]开关已经关闭,当前事件为=%s\n", tools.GetTimeNow(), data))
		log.Printf("开关已经关闭,当前事件为=%s\n", data)
		return
	}
	//赋值对象
	data.AddonRunners = job
	//取ipgroup组信息
	ipgroup, err := salt.GetCMDBInfo(data.Eventhand.Address)
	if err != nil {
		fmt.Println(err)
		ipgroup = data.Eventhand.Address
		//conf.WriteLog(fmt.Sprintf("%s[Return]重置minion-ip=%s\n", time.Now().Format("2006-01-02 15:04:05"), ipgroup))
		log.Printf("重置minion-ip=%s\n", ipgroup)
	}
	//进行机器的存活检测
	data.AddonRunners.Job.Tgt = activeaddress(ipgroup, data)
	//conf.WriteLog(fmt.Sprintf("%s[Return]CMDB返回的消息=%s,当前触发任务的IP=%s", time.Now().Format("2006-01-02 15:04:05"), data.AddonRunners.Job.Tgt, data.Eventhand.HostName+" "+data.Eventhand.Address))
	log.Printf("CMDB返回的消息=%s,当前触发任务的IP=%s\n", data.AddonRunners.Job.Tgt, data.Eventhand.HostName+" "+data.Eventhand.Address)
	//判断是否在队列中
	if err := reddao.SaddQueue("Eventlist", data.Eventhand.Address); err != nil {
		//conf.WriteLog(fmt.Sprintf("%s[DEBUG]事件已经加入到执行队列中该事件不能被重复执行=%s\n", tools.GetTimeNow(), err))
		log.Printf("事件已经加入到执行队列中该事件不能被重复执行=%s\n", err)
		return
	}
	//conf.WriteLog(fmt.Sprintf("%s[DEBUG]事件不存在执行队列中可以执行=%s\n", tools.GetTimeNow(), data.Eventhand.Address))
	log.Printf("事件不存在执行队列中可以执行=%s\n", data.Eventhand.Address)
	//进行Post请求取回事物执行ID
	resultid := salt.PostModulJob(info, &job.Job)
	//conf.WriteLog(fmt.Sprintf("%s[Return]resultid=%s\n", time.Now().Format("2006-01-02 15:04:05"), resultid))
	log.Printf("任务ID -> resultid=%s\n", resultid)
	//不存在则成立跳出循环
	if !reflect.ValueOf(resultid.Return[0].Minions).IsNil() && len(resultid.Return[0].Minions) > 0 {
		//构造对象
		data.JobReceipt = resultid
		//conf.WriteLog(fmt.Sprintf("%s[Return]异步任务返回的消息%s\n", time.Now().Format("2006-01-02 15:04:05"), data.JobReceipt.Return))
		log.Printf("开始执行异步任务返回的消息:%s,当前执行的IP为:%s\n", data.JobReceipt.Return, data.AddonRunners.Job.Tgt)
		conf.Chan2 <- data
		//排除空数组行为
	} else {
		//设置错误处理钉钉告警
		md := dd.SetDingError("执行任务错误请查看", data.Eventhand.Address, data.Notifications.CommonAnnotations["labels"], resultid.Return[0].Jid, data.Notifications.CommonAnnotations["description"], "无法获取到正确的minion-IP", active)
		conf.ChanDD <- md
	}
	//直接结束事件
	return
}

//执行jobs事件的查询
func handl(info *conf.AllMessage) {
	//统计查询的次数conf
	count := info.JobReceipt.Count
	count++
	//释放队列中的任务
	defer reddao.ZremDate("Eventlist", info.Eventhand.Address)
	//查询Token
	token := reddao.GetDate("token")
	//执行任务的ID号
	jid := info.JobReceipt.Return[0].Jid
	//超时中断指令
	if count == info.AddonRunners.Count && info.AddonRunners.TimeoutNUM == 0 {
		//存活检测
		_, state := salt.ActiveSalt(info.Eventhand.Address)
		active = state
		//写到数据库中
		reddao.SaddDate(info.JobReceipt.Return[0].Jid)
		//构造钉钉消息
		markdown := dd.SetDingError("执行任务超时请查看", info.Eventhand.Address, info.Eventhand.HostName, info.JobReceipt.Return[0].Jid, info.Notifications.CommonAnnotations["description"], "执行目标任务超时3分钟无回复", active)
		conf.ChanDD <- markdown
		//清理队列
		reddao.ZremDate("Eventlist", info.Eventhand.Address)
		//发送钉钉消息
		//conf.WriteLog(fmt.Sprintf("%s[Result]执行结果反馈 %s\n", time.Now().Format("2006-01-02 15:04:05"), info.JobReceipt.Return[0].Minions, "+", jid, "无法获取到JOb信息"))
		log.Printf("执行结果反馈 %s\n", info.JobReceipt.Return[0].Minions, "+", jid, "无法获取到JOb信息")
		return
	} else {
		if count == info.AddonRunners.Count {
			//抛弃旧的事件创建一个新的事件去处理
			info.AddonRunners.TimeoutNUM -= 1
			info.JobReceipt.Count = 0
			log.Printf("事件重新提交新任务,剩余提交次数为:%s", info.JobReceipt.Count)
			conf.Chan1 <- info
		}
	}
	//查询任务情况
	data := salt.QueryJob(jid, token)
	//排除空数组行为
	if reflect.ValueOf(data.Info[0].Minions).IsNil() {
		return
	}
	//判断是否取值成功,失败则重新进入队列等待再次的处理(返回消息不为空并且状态为真)
	if data.Info[0].Result[data.Info[0].Minions[0]].Success && data.Info[0].Result[data.Info[0].Minions[0]].Return != "" {
		//构造写入redis的数据信息
		endjob := conf.SetData(data)
		a, err := json.Marshal(endjob)
		if !tools.CheckERR(err, "构造写入redis的数据信息Json Manshal is Failed") {
			return
		}
		//记录时间
		info.TimeTotles.EndTime = time.Now()
		//总耗时时间
		info.TimeTotles.TotleTime = time.Since(info.TimeTotles.BeginTime)
		//构造钉钉消息
		markdown := dd.SetDD("日志处理结果请查看", info, endjob)
		//发送钉钉消息
		conf.ChanDD <- markdown
		//清理队列
		reddao.ZremDate("Eventlist", info.Eventhand.Address)
		//设置关闭
		info.AddonRunners.RetrySwitch = 1
		//写入redis数据库(data.Info[0].Result[key].Return)
		if err := reddao.InsertDate(jid, string(a)); err != nil {
			fmt.Println(err)
			return
		}
		//conf.WriteLog(fmt.Sprintf("%s[info]插入数据库成功=%s\n", time.Now().Format("2006-01-02 15:04:05"), info.JobReceipt.Return[0].Jid))
		log.Printf("插入数据库成功=%s\n", info.JobReceipt.Return[0].Jid)
		return
	} else {
		info.JobReceipt.Count = count
		if count%10 == 0 {
			log.Printf("[Process]没有获取到 节点=%s,次数=%s,ID=%s\n", info.JobReceipt.Return[0].Minions, info.JobReceipt.Count, info.JobReceipt.Return[0].Jid)
		}
		//进行异步任务的查询(默认180秒)
		//log.Printf("异步查询重置查询:%s,超时次数:%s",info.AddonRunners.Count,info.AddonRunners.TimeoutNUM)
		conf.Chan2 <- info
	}
	return
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
			//conf.WriteLog(fmt.Sprintf("%s[info]事件判断不成立=%s\n", a))
			log.Printf("[info]事件判断不成立=%s\n", a)
			return false
		}
	}
	//Hostname判断
	evhost := strings.Split(para.ParaHost, ",")
	for _, b := range evhost {
		if b != "" && strings.Contains(hostname, b) {
			//conf.WriteLog(fmt.Sprintf("%s[info]主机名判断不成立=%s\n", b))
			log.Printf("[info]主机名判断不成立=%s\n", b)
			return false
		}
	}
	return true
}

//存活检测
func activeaddress(ipgroup string, data *conf.AllMessage) (ip string) {
	ctx1 := context.Background()
	ctx, cancal := context.WithTimeout(ctx1, 60)
	defer cancal()
	data.Activechan = make(chan conf.Activestates, 5)
	//IP分割
	ipgroups := strings.Split(ipgroup, ",")
	//如果存在则进行存活检测否则直接执行
	//if len(ipgroup) >= 2 {
	//	for _, ip = range ipgroups {
	//		if ok, state := salt.ActiveSalt(ip); ok {
	//			active = state
	//			conf.WriteLog(fmt.Sprintf("%s[Return]最终存活的IP=%s\n,EerrorInfo=%s", time.Now().Format("2006-01-02 15:04:05"), ip, state))
	//			return ip
	//		} else {
	//			conf.WriteLog(fmt.Sprintf("%s[Return]检测失败的IP=%s\n,EerrorInfo=%s", time.Now().Format("2006-01-02 15:04:05"), ip, state))
	//		}
	//	}
	//} else {
	//	conf.WriteLog(fmt.Sprintf("%s[Return]只存在一个IP不需要存活检测=%s\n", time.Now().Format("2006-01-02 15:04:05"), ipgroup))
	//	return ipgroup
	//}
	//进行检测IP的存活情况
	if len(ipgroups) > 1 {
		for _, ip := range ipgroups {
			go salt.ActiveSalttest(ctx, ip, data)
			log.Printf("开始调用存活检查IP为=%s", ip)
		}
		//如果为true则跳出循环切调用cancal(没有循环取是因为存活判断,活返回的时间一定是大于死的超时时间所以不去select取)
		for i := 0; i != len(ipgroups); i++ {
			result := <-data.Activechan
			log.Printf("返回的结果为=%s", result)
			//正常存活返回
			if result.States {
				active = result.Msg
				return result.Address
			}
		}
	}
	//一切异常都取值第一个去执行
	return ipgroups[0]
}

//提供接口调用salt
func ApiSalt(obj *conf.JobRunner) {
	//构造空的对象(满足上需求)
	var dataobj *conf.AllMessage
	//获取Token信息
	token, err := salt.Check()
	tools.CheckERR(err, "取回token失败")
	//获取auth Token
	if auth := reddao.GetDate("AuthToken"); auth == "" {
		err := salt.GetCMDBAUTH()
		if err != nil {
			//conf.WriteLog(fmt.Sprintf("[error]无法获取到token,%s\n", err))
			log.Printf("无法获取到token,%s\n", err)
			return
		}
	}
	//取ipgroup组信息
	ipgroup, err := salt.GetCMDBInfo(obj.Tgt)
	if err != nil {
		fmt.Println(err)
		ipgroup = obj.Tgt
		//conf.WriteLog(fmt.Sprintf("%s[Return]重置minion-ip=%s\n", time.Now().Format("2006-01-02 15:04:05"), ipgroup))
		log.Printf("重置minion-ip=%s\n", ipgroup)
	}
	//存活检测
	obj.Tgt = activeaddress(ipgroup, dataobj)
	//post salt-master API Return Result
	result := salt.PostRsyncModulJob(token, obj)
	conf.ChanResult <- result
}
