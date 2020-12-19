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
$ ./apollo-agent -c app-example.yaml
```
Tips：

1、请正确配置 app-example.yaml文件中关键内容

2、生产环境请使用systemd或supervisor，常驻agent进程

### 配置文件说明
以app-example.yaml为例
```yaml
client:               # agent本地配置信息
  pollOrWatch: watch  # 拉取配置的方式，支持poll和watch
  allInOne: true      # 拉取的配置数据是否需要合并到一个文件
  ip: 127.0.0.1       # 获取灰度版本的client ip
  logExpire: 72h      # agent本地日志的过期时间，过期自动清理防止日志过多
  beatFreq: 2s        # agent 心跳频率，该配置值不支持热更新

server:               # Apollo Config Service相关信息
  address: http://your-apollo.config-service.address # Config Service地址
  cluster: default    # 集群

apps:                 # Apollo的应用列表
  - appId: demo       # Apollo上的应用appId
    secret: a93ab23   # 如果应用开启了访问认证，需要配置访问密钥
    namespace:        # 应用下的Namespace信息，需要写上详细的类别后缀
      - application.properties
      - redis.properties
      - mysql.properties
    pollInterval: 2s  # 如果agent拉取为poll方式，poll的周期值，watch方式，此配置无作用
    inOne:            # 如果agent拉起配置合并到一个文件，即client.allInOne = true，指定了合并后文件的信息（文件名及文件内容格式）
      filename: ./demo.env
      syntax: env     # 仅支持 dotEnv、ini(非严格env和ini，仅key=value对)、php、txt(包含yaml、yml、json、txt)
```
以上所有配置项，除client.beatFreq不支持热更新（直接修改保存即生效，不需重启服务），其他均支持热更新，良好的处理了agent进程无重启权限的问题。