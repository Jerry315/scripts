1. 环境初始化

   ```bash
   pip install -r requirements.txt
   ```

2. 规则说明

   - 一分钟内群发或对单个用户最多可以发送30条信息
   - 一分钟内超过30条的信息会被保留到下一个一分钟发送

3. 安装redis

4. 配置说明

   ```python
   config.yml
   wechat: # 微信相关信息配置
     corpid: xxxxx # 微信企业id
     token_url: https://qyapi.weixin.qq.com/cgi-bin/gettoken
     send_url: https://qyapi.weixin.qq.com/cgi-bin/message/send
   redis: # redis信息配置
     host: 192.168.2.73
     passwd: 123456
     port: 6800
     db: 4
   smtp: # 邮件信息配置
     subject: xxxxx
     user: xxxxx@topvdn.com
     passwd: xxxxxx!
     smtp_server: xxxx.topvdn.com
     receivers:
       - xxxxx@topvdn.com
   alert: # 告警信息
     Warning:
       toparty: 6|7
       agentid: 15
       wechat_corpsecret: xxxx
       P: P4
     Average:
       toparty: 6|7
       agentid: 20
       wechat_corpsecret: xxxx
       P: P3
     High:
       toparty: 2
       agentid: 1
       wechat_corpsecret: xxxx
       P: P2
     Disaster:
       toparty: 2
       agentid: 11
       wechat_corpsecret: xxxx
       P: P1
     Other:
       toparty: 6|7
       agentid: 15
       wechat_corpsecret: xxxx
       P: P4
   log_dir: /tmp/sendwechat #日志路径
   record_file_dir: /tmp/sendwechat/dial # 记录未恢复报警信息目录
   alert_url: http://42.51.12.155:8008/
   log_dir: /tmp/sendwechat # 日志目录
   record_file: /tmp/sendwechat/P1.txt # 记录未恢复报警信息，需要手动创建，文件授权zabbix用户组
       
   toparty： 部门编号，多个部门用|隔开
   agentid：应用id
   wechat_corpsecret： 应用secret
   P：定义告警级别
   
   ```


