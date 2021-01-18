### v4.1.4-release
1、支持将apolloAgentForPHP配置文件直接转换为当前版本可使用（poll模式），转换命令
```shell script
./apollo-agent -convertConfig <oldConfigFile> <newConfigFile>
```
`oldConfigFile可缺省，默认为：/opt/app/apolloAgentForPHP/conf/app.yaml`

`newConfigFile可缺省，默认为：/opt/app/apollo-agent/conf/app.yaml`

`如果使用2345-rpm安装，可执行 /opt/app/apollo-agent/bin/apollo-agent -convertConfig 后，直接
执行 systemctl restart 2345-apollo-agent，即可启动新版本`

2、因watch模式存在请求超时情况，还未定位详细原因，暂时关闭此模式，同时，poll模式默认轮询周期调整为15秒一次。

### v4.2.0
1、修复了watch模式的超时问题，重新开启了watch模式。