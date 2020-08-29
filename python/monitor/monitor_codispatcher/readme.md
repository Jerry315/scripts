## codispatcher监控脚本

参数说明：

- -q，查询codispatcher配置文件中监听的端口；
- -p或--port，查询port是否在监听。

配置说明：

```
配置文件config.yml
config:
  codispatcher:
    conf_dir: '/opt/codispatcher/etc' # codispatcher的配置文件位置
```

查询规则：

- 查找/opt/codispatcher/etc目录下所有以.yaml结尾的配置文件，并读取文件，返回配置文件以及监听端口；
- 根据端口和应用名称，返回对应的pid。

使用说明：

```
返回配置文件和端口
./monitor_codispatcher.sh 
{"data": [{"{#DEBUGPORT}": "6030", "{#QUERYPORT}": "8081", "{#CONF}": "/opt/codispatcher/etc/436207622.codispatcher.yaml"}]}

根据port返回进程的pid
./monitor_codispatcher.sh 8081
11584
如果返回0，则异常
```

