## 统计邮件汇总

#### 功能说明

将cid存储周期一致性校验和cid超时未推流或未上传图片的检查结果统一处理，然后根据类别汇总后，根据类别发送汇总邮件

#### 配置说明

##### agent

```yaml
# converge_alert_mail_agent.yaml
limit: # 
  message_timestamp: 350000 # 中央数据中的字段，判断cid是否在线，单位（纳秒）
  step: 2000 # 调用checkserver接口每次最多携带多少个cid
  timeout: 2 # 用于获取转发上推流时长大于多长时间，单位（分钟）
log:
  filename: converge_alert_mail_agent.log
  format: json
  level: info
  layout: 20060102
  path: log
  cycleFile: cycleFile
  timeOutFile: timeOutFile
  expire: 7
mongodb: # 中央数据库配置
  camera:
    db: camera
    fields:
      - _id
      - message_timestamp
      - push_state
    table: device
    url: mongodb://camera:xxxxx@monghost:27017/?authSource=camera&authMechanism=SCRAM-SHA-1
  devices: # 通配设备数据库配置
    db: devices
    fields:
      - _id
      - storage
      - pic_storage
      - sn
      - name
      - brand
      - model
      - software_version
      - software_build
    table: devices_camera
    url: mongodb://devices:xxxxx@monghost:27017/?authSource=devices&authMechanism=SCRAM-SHA-1
  mawarapp: # 通配app数据库配置
    db: mawarapp
    fields:
      - _id
      - group
    table: mawarapp_camera
    url: mongodb://mawarapp:xxxxx@monghost:27017/?authSource=mawarapp&authMechanism=SCRAM-SHA-1
checkServer: http://127.0.0.1:7061
whitelist: # 白名单功能
  all: null # 存储周期一致性，全局cid白名单
  picture: null # 图片cid存储周期白名单
  video: null # 视频cid存储周期白名单
  relay: # cid超时未推流和上传图片白名单
relay: # 转发地址相关配置，如果有开启HTTPbasic auth需要提供用户名和密码
  urls: null
  username: ""
  password: ""
timeout: 86400 # 统计超时多久的cid未推流和上传图片，单位（秒）
project: dgdx 
zname: 东莞云 
secretid: jerry # 要和server端保持一致
secretkey: 123456
server: http://127.0.0.1:1212 # server端监听地址
```

##### server

```yaml
bind: 127.0.0.1 
port: 5050
deviceCycle: # 除subject可修改，其他修改慎重
  name: deviceCycle
  template: deviceCycle
  subject: 存储周期每日全量检测
deviceTimeOut:
  name: deviceTimeOut
  template: deviceTimeOut
  subject: cid超过24个小时未上传图片和推流
limit: 10
log:
  exfile: cid.xlsx
  expire: 7
  format: json
  layout: 20060102
  level: info
  name: converge_alert_mail_server.log
  path: log
mail:
  recive:
    admin:
      - zhanlin@antelope.cloud
    normal:
      - zhanlin@antelope.cloud
  send:
    from: xxxxx@topvdn.com
    password: xxxxxx
    server: smtp.mxhichina.com # 使用公司邮件，不需要修改这个配置
    username: xxxxx@topvdn.com
mongodb: # 本地存储数据的数据库
  db: report
  table: device
  url: mongodb://report:xxxxx@monghost:27017/?authSource=report&authMechanism=SCRAM-SHA-1
secretid: jerry
secretkey: 123456

```

#### 规则说明

##### agent采集

```
存储周期一致性
步骤1 定时任务执行脚本，查询中央数据库，获取心跳时间(字段名message_timestamp)比当前时间差小于350s和状态(字段名push_state)为4的记录，记录包含cid；
步骤2；查询通配数据库，获取记录包含cid、视频周期、图片周期，并与'步骤1'中的cid对比，取cid交集的记录，过滤出白名单中的cid；
步骤3 根据'步骤2'得到的cid列列表，循环调用对象存储'checkserver'的接口，获取最新的视频周期和图片周期；
步骤4 根据'步骤2'和'步骤3'返回的结果进行比对，同一个cid在通配的视频(图片)的存储周期与对象存储的视频（图⽚）的存储周期不一致则认定异常结果，若'checkserver'接口返回的cid的time字段为0，cycle为-1，该种情况（cid从未推流或者上传图片）记录为异常周期cycle记为'-1'。异常结果记录对应的cid、通配的视频周期、通配的图片周期、对象存储的视频周期、对象存储的图片周期，发送到server端。

cid超时未推流和未上传图片
步骤1 通过转发地址，获取所有cid，所有cid满足推流大小(bw_in > 100kb)并且推流时间(time >= 2分钟)，并对cid去重；
步骤2 根据配置文件中step大小，每次取step个cid进行检查，将返回结果放在列表中，直到所有cid都检查完毕，统计出所有cid数量，以及满足规则说明中的规则2的cid数量，结束循环；
步骤3 根据checkmode，如果是0对步骤2中所有的cid根据LatestImageTime进行从小到大排序，如果是1对步骤2中所有的cid根据LatestVideoTime进行从小到大排序；
步骤4 将统计结果发送到server端。
```

#### 使用说明

```bash
# agent 计划任务每天六点半上报数据
30 6 * * * root /opt/sa_tools/scripts/go/converge_alert_mail_agent/converge_alert_mail_agent

# server 计划任务每天九点一刻发送报告
15 9 * * * root /opt/sa_tools/scripts/converge_alert_mail_server/converge_alert_mail_server dcr
15 9 * * * root /opt/sa_tools/scripts/converge_alert_mail_server/converge_alert_mail_server dtr
```



