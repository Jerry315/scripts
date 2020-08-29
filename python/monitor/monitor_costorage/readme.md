## 磁盘lld监控脚本

参数说明：

- -q，查询storage.conf配置文件中partition的enable为1的磁盘upload_port；
- -p或--port，查询port是否在监听。

配置说明：

```
配置文件config.yml
config:
  costorage:
    config: '/opt/COStorage/storage.conf' #costorage的配置文件位置
```

查询规则：

- storage.conf配置文件中partition的enable为1的partition；
- 统计enable为1的partition数量。

使用说明：

```
查询partition的upload_port
python monitor_costorage.py -q
{'{#ONLINE}': 12, 'data': [{'{#UPLOAD_PORT}': '9101'}, {'{#UPLOAD_PORT}': '9102'}, {'{#UPLOAD_PORT}': '9103'}, {'{#UPLOAD_PORT}': '9104'}, {'{#UPLOAD_PORT}': '9105'}, {'{#UPLOAD_PORT}': '9106'}, {'{#UPLOAD_PORT}': '9107'}, {'{#UPLOAD_PORT}': '9108'}, {'{#UPLOAD_PORT}': '9109'}, {'{#UPLOAD_PORT}': '9110'}, {'{#UPLOAD_PORT}': '9111'}, {'{#UPLOAD_PORT}': '9112'}]}


查询9102端口是否存在
python monitor_costorage.py -p 9102
1
1，存在，0不存在
```

