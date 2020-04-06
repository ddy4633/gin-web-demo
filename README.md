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