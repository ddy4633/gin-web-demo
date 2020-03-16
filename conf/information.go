package conf

var (
	Chan1 = make(chan *EventHand, 0)
	Chan2 = make(chan *JobReturn, 10)
	Token = ""
)

//封装Ansible
type AnsibleAPI struct {
	AddrIP []string //调用的IP
	Module string   //调用的模块
	Shell  string   //调用的命令
}

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

//具体返回信息
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
type Returninfo struct {
	Return []Info `json:"return"`
}

//salt基础信息
type SaltInfo struct {
	Token string `json:"token"`
	User  string `json:"user"`
	Eauth string `json:"eauth"`
}

//返回的Job信息
type JobReturn struct {
	Return []List `json:"return"`
	Count  int
}
type List struct {
	Jid     string   `json:"jid"`
	Minions []string `json:"minions"`
}

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
	Client string `json:"client"`
	//minions机器名称
	Tgt string `json:"tgt"`
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
	Expr_form string `json:"expr_form"`
	//func执行函数
	Fun string `json:"fun"`
	//fun的参数项
	Arg string `json:"arg"`
}

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
	JobInfo struct {
		Info   []JobMessage `json:"info"`
		Return []string     `json:"return"`
	}
	JobMessage struct {
		Function  string            `json:"Function"`
		Minions   []string          `json:"Minions"`
		Result    map[string]string `json:"Result"`
		StartTime string            `json:"StartTime"`
	}
)

//常量值
const (
	Json_Accept       = "application/json"
	Json_Content_Type = "application/json"
	URL               = "http://10.200.10.23:8800/"
	URL_job           = "http://10.200.10.23:8800/jobs/"
)
