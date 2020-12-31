## apollo-agent

Apollo Go Agent

[![GitHub Release](https://img.shields.io/github/release/2345tech/apollo-agent.svg)](https://github.com/2345tech/apollo-agent/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


### 功能介绍
1、将Apollo配置中心的数据拉取到本地文件，支持同时拉取多个应用配置数据

2、支持Apollo Config Service、AppId、Cluster、Namespace等信息热更新

### 使用场景
1、PHP-FPM项目（如：laravel）依赖Apollo配置中心进行配置管理，支持Agent作为旁路系统拉取配置到指定文件

2、非业务场景下（如：CI/CD过程）依赖Apollo配置中心进行配置管理，支持Agent作为旁路系统拉取配置到指定文件

### 使用说明
```shell script
$ go clone https://github.com/2345tech/apollo-agent.git
$ cd apollo-agent
$ go build
$ ./apollo-agent -h
$ ./apollo-agent -c app-example.yaml -l agent.log
```
Tips：

1、请正确配置 app-example.yaml文件中关键内容

2、生产环境请使用systemd或supervisor，常驻agent进程

### 配置文件说明
以app-example.yaml为例
```yaml
client: # agent本地配置信息
  pollOrWatch: watch  # 拉取配置的方式，支持poll和watch
  allInOne: false     # 拉取的配置数据是否需要合并到一个文件
  ip: 127.0.0.1       # 获取灰度版本的client ip
  logExpire: 72h      # agent本地日志的过期时间，过期自动清理防止日志过多
  beatFreq: 2s        # agent 心跳频率，该配置值不支持热更新，不配置默认为10m(分钟)

server: # Apollo Config Service相关信息
  address: http://your-apollo.config-service.address # 指定环境的Config Service地址
  cluster: default    # 集群名称

apps: # Apollo的应用列表
  - appId: demo       # Apollo上的应用appId
    secret: a93ab23   # 如果应用开启了访问认证，需要配置访问密钥
    namespace: # 应用下的Namespace信息，当非properties类别的NS时，必须要写上详细的类别后缀
      - application.properties
      - redis.json
      - mysql.yaml
    pollInterval: 2s  # 如果agent拉取为poll方式，poll的周期值，watch方式，此配置无作用
    syntax: env       # 仅支持 dotEnv、ini(非严格env和ini，仅key=value对)、php、txt(包含yaml、yml、json、txt)
    inOneFile: ./.env # 如果agent拉起配置合并到一个文件，即client.allInOne = true，指定了合并后文件的信息（文件名及文件内容格式）
    # 当client.allInOne = false，会为每个namespace生成一个独立的文件（目录位置与inOneFile相同），如上：./application.properties、./redis.json、./mysql.yaml
```
以上所有配置项，除client.beatFreq不支持热更新（直接修改保存即生效，不需重启服务），其他均支持热更新，良好的处理了agent进程无重启权限的问题。

### 容器部署
可将agent作为应用容器的sidecar部署，此部署方式推荐使用环境变量作为启动配置（非容器也支持环境变量作为启动配置）

当启用环境变量作为启动配置时，所有agent log 会写入到 /dev/stdout

| 环境变量名称 | 默认值 | 说明 |
|------------|-------|-----|
| APOLLO_AGENT_CLIENT_TYPE | poll | agent拉取配置方式，默认使用poll |
| APOLLO_AGENT_CLIENT_ALLINONE | true | 默认拉取配置后会合并到一个文件 |
| APOLLO_AGENT_CLIENT_LOGEXPIRE | 24h | 默认agent本地日志文件保留1天，注意是一个自然天，不是24小时，且最小单位天 |
| APOLLO_AGENT_CLIENT_IP | 空字符串 | 默认不配置灰度版本ip |
| APOLLO_AGENT_CLIENT_BEATFREQ | 10m | 默认agent会10分钟记录一次心跳日志 |
| APOLLO_AGENT_SERVER_ADDRESS | 空字符串 | apollo config service地址 |
| APOLLO_AGENT_SERVER_CLUSTER | default | 默认拉取当前环境的default集群配置 |
| APOLLO_AGENT_APP_ID | 空字符串 | 需要拉取配置的appId |
| APOLLO_AGENT_APP_NAMESPACES | application.properties | 默认拉取application.properties，如果有多个请使用,号隔开 |
| APOLLO_AGENT_APP_SECRET | 空字符串 | 访问密钥 |
| APOLLO_AGENT_APP_SYNTAX | env | 如果拉取配置后会合并到一个文件，合并后文件默认类型是dotEnv |
| APOLLO_AGENT_APP_POLL_INTERVAL | 60s | 如果是poll方式，默认的interval为60秒 |
| APOLLO_AGENT_APP_IN_ONE_FILE | ./application.properties | 如果开启allInOne，默认拉取配置后会合并到application.properties |

注意：使用环境变量启动agent，只支持拉取一个appId，如果需要拉取多个，请使用配置文件方式启动