### 爱耳目事件报警条数统计

环境安装：

```
分别进入agent和server目录执行：./install.sh
```

启动：

```
agent
	进入agent目录：nohup ./start.sh >> log/stdout.log
server
	添加30 0 * * * root /bin/bash /opt/sa_tools/IermuStats/server/start.sh到/etc/crontab
```

配置：

```
agent配置文件config.yml
config:
  redis:
    startup_nodes: # redis集群信息，host主机，port端口，password密码
      - host: 192.168.2.25
        port: 6377
      - host: 192.168.2.25
        port: 6378
    password: 123456
  host: 127.0.0.1 # agent启动绑定host，
  port: 5000 # agent启动运行的端口
  
server配置文件config.yml
config:
  stat_url: # agent的url
    - http://127.0.0.1:5000/v1/iermu/stats
    - http://127.0.0.1:5000/v1/iermu/stats
  limit: 30 # 统计报警事件条数
  smtp:
    subject: 【数据统计】爱耳目事件报警条数统计
    user: xxxxx@topvdn.com
    passwd: xxxxxx!
    smtp_server: xxxx.topvdn.com
    receivers:
      - xxxxx@topvdn.com
```

