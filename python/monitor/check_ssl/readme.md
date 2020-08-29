## ssl证书到期检测脚本

环境初始化
```
./install.sh
```

参数说明：

```bash
-d或--dns：解析域名后端真实主机ip列表
-e或--expire：预过期天数
-u或--url: 被检查的域名
-i或--ip：域名实际绑定的主机
-p或--port：域名监听的端口，默认443
-q或--query：手动执行脚本查询域名证书到期与否
-z或--zabbix：zabbix调用脚本时需要传的参数
```

使用方法：

```
方式一：
	所有信息从配置文件config.yml中读取
	 ./start.sh

方式二：
	手动执行脚本查询指定域名
	./start.sh -u www.xxxx.cn -i 1.1.1.1 -q
	
方式三zabbix调用：
    第一步:解析域名，获取ip和port
    	 ./start.sh -q
    	 {"data": [{"{#IP}": "1.1.1.1", "{#PORT}": 443 ,"{#URL}": "www.xxx.com"}, {"{#IP}": "2.2.2.2", "{#PORT}": 443,"{#URL}": "www.xxx.com"}, {"{#IP}": "3.3.3.3", "{#PORT}": 443,"{#URL}": "www.xxx.com"}]}
    第二步：查询对应ip和域名的证书到期时间，返回实际到期天数
    	 ./start.sh -u www.xxxx.c -i 3.3.3.3 -p 443 -z
    	 669
```

