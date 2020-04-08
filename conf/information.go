package conf

import (
	"fmt"
	"gin-web-demo/tools"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	//告警结构构造
	Chan1 = make(chan *AllMessage, 100)
	//事务处理事件拉取
	//Chan2 = make(chan *JobReturn, 30)
	Chan2 = make(chan *AllMessage, 30)
	//钉钉通知
	ChanDD = make(chan *DingTalkMarkdown, 15)
	//全局配置信息
	Config Ginconf
)

//封装Ansible命令
type AnsibleAPI struct {
	AddrIP []string //调用的IP
	Module string   //调用的模块
	Shell  string   //调用的命令
}

//封装ansible-palybooks

//封装SSH
type SSHInfo struct {
	Addr     string
	User     string
	Password string
	Port     string
}

//处理事件结构体
type EventHand struct {
	EventName string `json:"event_name"`
	HostName  string `json:"host_name"`
	Address   string `json:"address"`
	Event     string `json:"event"`
	Status    int    `json:"status"`
}

//事件处理
type EventsWait struct {
	//通道类型定义
	cha <-chan EventHand
	//暂时留给锁的位置
}

//Token具体返回信息
type Info struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Eauth     string   `json:"eauth"`
	StartTime float64  `json:"start"`
	EndTime   float64  `json:"expire"`
	Perms     []string `json:"perms"`
	Token     string   `json:"token"`
}

//返回信息
type (
	Returninfo struct {
		Return []Info `json:"return"`
	}
	//salt基础信息
	SaltInfo struct {
		Token string `json:"token"`
		User  string `json:"user"`
		Eauth string `json:"eauth"`
	}
	//返回的Job信息
	JobReturn struct {
		Return []List `json:"return"`
		Count  int    `json:-`
	}
	List struct {
		Jid     string   `json:"jid"`
		Minions []string `json:"minions"`
	}
)

//任务结构体
type JobRunner struct {
	/*
		 local : 使用‘LocalClient <salt.client.LocalClient>’ 发送命令给受控主机，等价于saltstack命令行中的'salt'命令
		 local_async : 和local不同之处在于，这个模块是用于异步操作的，即在master端执行命令后返回的是一个jobid，任务放在后台运行，通过产看jobid的结果来获取命令的执行结果。
		 runner : 使用'RunnerClient<salt.runner.RunnerClient>' 调用salt-master上的runner模块，等价于saltstack命令行中的'salt-run'命令
		 runner_async : 异步执行runner模块
		 wheel : 使用'WheelClient<salt.wheel.WheelClient>', 调用salt-master上的wheel模块，wheel模块没有在命令行端等价的模块，
		但它通常管理主机资源，比如文件状态，pillar文件，salt配置文件，以及关键模块<salt.wheel.key>功能类似于命令行中的salt-key。
	*/
	//模块名称
	Client string `json:"client,omitempty"`
	//minions机器名称
	Tgt string `json:"tgt,omitempty"`
	/*
		'glob' - Bash glob completion - Default
		'pcre' - Perl style regular expression
		'list' - Python list of hosts
		'grain' - Match based on a grain comparison
		'grain_pcre' - Grain comparison with a regex
		'pillar' - Pillar data comparison
		'nodegroup' - Match on nodegroup
		'range' - Use a Range server for matching
		'compound' - Pass a compound match string
	*/
	//对tgt的匹配规则
	Expr_form string `json:"expr_form,omitempty"`
	//func执行函数
	Fun string `json:"fun,omitempty"`
	//fun的参数项
	Arg string `json:"arg,omitempty"`
	//要过滤的参数选项
	//Para string `json:"-"`
}

//带过滤的事件
type (
	AddonJobRunner struct {
		Job JobRunner
		//要过滤的参数选项
		Para ParaMeter `json:"para"`
		//是否执行处理
		Switch int `json:"switch"`
		//任务超时的次数
		Count int `json:"count"`
	}
	ParaMeter struct {
		//主机名过滤
		ParaHost string `json:"parahost"`
		//过滤事件
		ParaEvent string `json:"paraevent"`
	}
)

//自定义需要接收AlterManager的结构体
type (
	HookMessageInfo struct {
		//定义信息状态
		Status string `json:"status"`
		//标签信息
		Alerts []Alert `json:"alerts"`
		//命令描述信息
		CommonAnnotations map[string]interface{} `json:"commonannotations"`
		//回调的web-hook信息
		Receiver string `json:"receiver"`
		//告警基础信息
		CommonLabels map[string]interface{} `json:"commonLabels"`
	}

	Alert struct {
		//描述信息
		Annotations map[string]string `json:"annotations"`
		//开始时间
		StartsAt string `json:"startsat"`
		//结束时间
		EndsAt string `json:"endsat"`
		//标签信息
		Labels map[string]string `json:"labels"`
		//查看信息URL
		GeneratorURL string `json:"generatorurl"`
		//状态
		Status string `json:"status"`
	}

	ComMon struct {
		Alertname string `json:"alertname"`
		Hostname  string `json:"hostname"`
		Instance  string `json:"instance"`
		Team      string `json:"team"`
	}
)

//返回的异步任务信息
type (
	DDMsg struct {
		Info  JobInfo `json:"info"`
		Event string  `json:"event"`
	}
	JobInfo struct {
		Info     []JobMessage `json:"info"`
		Return   []string     `json:"return"`
		Hostname string       `json:"hostname"`
	}
	JobMessage struct {
		Jid       string          `json:"jid"`
		Function  string          `json:"Function"`
		Minions   []string        `json:"Minions"`
		Result    map[string]Data `json:"Result"`
		StartTime interface{}     `json:"StartTime"`
		Arguments []string        `json:"Arguments"`
	}
	Data struct {
		Return  string `json:"return"`
		Retcode int    `json:"retcode"`
		Success bool   `json:"success"`
	}
)

//返回查询事件的参数
type EndJob struct {
	//任务开始的时间
	StartTime string `json:"start_time"`
	//任务总的执行完成时间
	AggrTime string `json:"aggr_time"`
	//目标主机
	Target []string `json:"target"`
	//执行的结果
	Info string `json:"info"`
	//对象的Jid
	Jid string `json:"jid"`
	//主机名
	Hostname string `json:"hostname"`
}

//使用钉钉的参数
type (
	Dingding struct {
		Token   string
		Message map[string]Dmessage
	}
	Dmessage struct {
		Msgtype  string
		markdown map[string]Detial
		At       *At `json:at`
	}
	Detial struct {
		Title string
		Text  string
	}
	At struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	}
)

//钉钉返回数据
type (
	//钉钉返回信息处理
	ReturnDD struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	//定义markdown格式
	DingTalkMarkdown struct {
		MsgType  string `json:"msgtype"`
		At       *At1
		Markdown *Markdown `json:"markdown"`
	}
	//dingd详细内容
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
	//通知的消息
	At1 struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	}
)

//测试用例
type AClert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:annotations`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

type Notification struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:receiver`
	GroupLabels       map[string]string `json:groupLabels`
	CommonLabels      map[string]string `json:commonLabels`
	CommonAnnotations map[string]string `json:commonAnnotations`
	ExternalURL       string            `json:externalURL`
	Alerts            []AClert          `json:alerts`
}

//sls格式返回数据的处理接收
type RetuenSls struct {
	Info []struct {
	} `json:"info"`
	Return []struct {
		IP struct {
			Cmd__mycommand__df__h__run struct {
				ID      string `json:"__id__"`
				RunNum  int64  `json:"__run_num__"`
				Sls     string `json:"__sls__"`
				Changes struct {
					Pid     int64  `json:"pid"`
					Retcode int64  `json:"retcode"`
					Stderr  string `json:"stderr"`
					Stdout  string `json:"stdout"`
				} `json:"changes"`
				Comment   string  `json:"comment"`
				Duration  float64 `json:"duration"`
				Name      string  `json:"name"`
				Result    bool    `json:"result"`
				StartTime string  `json:"start_time"`
			} `json:"cmd_|-mycommand_|-df -h_|-run"`
		} `json:"192.168.3.138"`
	} `json:"return"`
}

//大全局使用的结构体
type AllMessage struct {
	//需要处理的过滤事件集合
	AddonRunners *AddonJobRunner
	//接受AlterManager的webhook告警信息
	Notifications Notification
	//异步取执行的事务信息
	DDMsgs DDMsg
	//事件唯一ID号
	ID string
	//处理事件结构体
	Eventhand *EventHand
	//post的任务的异步回执单
	JobReceipt *JobReturn
}

//cmdb返回的IP
type Retuencmdb struct {
	Msg  string `json:"message"`
	Data struct {
		IPgroup  string `json:"ipgroup"`
		Hostname string `json:"hostname"`
	}
	Code int `json:"code"`
}

//获取cmdbtoken
type TokenCmdb struct {
	Token    string     `json:"token"`
	AuthCmdb AllMessage `json:"authcmdb"`
}

//认证CMDB的结构体
type AuthCmdb struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
}

//常量值
const (
	Json_Accept       = "application/json"
	Json_Content_Type = "application/json"
)

//返回构造好的插入redis中的结果数据
func SetData(data JobInfo) *EndJob {
	redata := &EndJob{
		StartTime: data.Info[0].StartTime.(string),
		AggrTime:  tools.GetTimeNow(),
		Target:    data.Info[0].Minions,
		Info:      data.Info[0].Result[data.Info[0].Minions[0]].Return,
	}
	return redata
}

//写日志信息
func WriteLog(obj string) {
	if Config.Conf.LogMod == "debug" || Config.Conf.LogMod == "" {
		fmt.Println(gin.DefaultWriter, obj)
	}
}
