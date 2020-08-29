## codproducer监控脚本

参数说明：

- -q，查询codproducer配置文件中监听的端口；
- -p或--port，查询port是否在监听。

配置说明：

```
配置文件config.yml
config:
  codproducer:
    conf_dir: '/opt/codproducer/etc'  #codproducer的配置文件位置
```

查询规则：

- 查找/opt/codproducer/etc目录下所有以.yaml结尾的配置文件，并读取文件，返回配置文件以及监听端口；
- 根据端口和应用名称，返回对应的pid。

使用说明：

```
返回配置文件和端口
./monitor_codproducer.sh 
{"data": [{"{#RPCPORT}": "8130", "{#QUERYPORT}": "8180", "{#CONF}": "/opt/codproducer/etc/A.codproducer.yaml"}, {"{#RPCPORT}": "8131", "{#QUERYPORT}": "8181", "{#CONF}": "/opt/codproducer/etc/B.codproducer.yaml"}]}

根据port返回进程的pid
./monitor_codproducer.sh 8180
4043
如果返回0，则异
```

