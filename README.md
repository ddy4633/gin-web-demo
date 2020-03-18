# gin-Fault-self-healing-demo

# 顾名思义

- webhook前置条件

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

- 实现的代码逻辑

![](https://i.imgur.com/WtuwuBX.png)