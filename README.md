# gin-Fault-self-healing-demo

# 顾名思义

>实现altermanager告警通过salt去调用定义的方法执行处理对应的告警,然后钉钉返回消息

- 已经完成`处理日志过大`

- 待完成`cpu过高分析`

- 待完成`MEM过高分析`

- 待完成`报表数据汇总管理展示`

- 待完成`提供通用的resterful-API`

## webhook前置条件

```shell script
#1.vim altermanager.yaml
#配置调用webhook
#golang-webhook调用例子配置告警的接受组
- name: 'golang_webhook'
  webhook_configs:
  - url: 'http://xxx:9090/monitor'
    send_resolved: true
 
#配置prometheus的rule报警规则
groups:
- name: golangwebhookAlert
  rules:
  - alert: gocpuUsageAlert
    expr: (100 - (avg by (instance,hostname)(irate(node_cpu_seconds_total{job="consul_sd_node_exporter",mode="idle"}[1m])) * 100)) >20
    for: 10s
    labels:
      team: golang-webhook

```

- 实现的代码逻辑架构图

![](https://i.imgur.com/WtuwuBX.png)

- 配置文件修改方式

```shell script
请求头: Content-Type:application/x-www-form-urlencoded
请求地址: http://你配置文件中写入的地址/config
请求参数:
client:saltstack执行方式
tgt:目标节点
func:执行的函数名称
arg:具体的命令
paraevent:过滤的hostname中包含的字段
switch:是否处理任务的开关(0开1关)
count:saltstack任务执行超时时间
parahost:过滤的instace信息
```

![](https://s1.ax1x.com/2020/04/08/GWwnFx.png)

## 运行方式

- 下载代码

- 执行go构建CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o gin-web main.go

- 执行Images构建：docker build -t xxxx/xxx/gin-web:v1 .

- 运行：docker run -itd -p 9090:9090 --name gin-web xxxx/xxx/gin-web:v1