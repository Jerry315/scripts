## 视频或图⽚片的存储周期监控

### 功能说明

#### 抽样对比

采集间隔：2min
报警手段：zabbix
采集规则：

```
步骤1 zabbix触发执行脚本时，检查上次是否有查询异常的cid存在，存在则直接使用前一次异常的cid ，跳到'步骤4'；
步骤2 通过请求转发的/stat接口获取数据，对数据处理，过滤出推流(bw_in > 100kb)并且推流时间(time > 2min 且 time <= 10min)的 cid 列列表和推流(bw_in > 100kb)并且推流时间(time > 10min)的 cid 列表；
步骤3 从上述满足条件的数据，过滤出白名单中的cid，在剩余数据中随机抽选16个cid进行检查(2-10min占比6成取整，>10min占比4成取整，数量不够比例例以实际数量为准)，且cid⼀定存在通配数据库中；
步骤4 根据 cid 查询通配数据库，获取对应的视频和图片的存储周期配置；
步骤5 根据 cid 调用对象存储'checkserver'接口，获取对应的视频和图片的实际存储的周期；
步骤6 根据'步骤4'和'步骤5'返回的结果进行比对，同一个cid在通配的视频(图片)的存储周期与对象存储的视频（图片）的存储周期不一致则认定异常结果，若'checkserver'接口返回的cid的time字段为0，cycle为-1，该种情况（cid从未推流或者上传图片）也认定为正常。正常结果则返回0，异常结果则记录日志并返回1；
步骤7 脚本任意环节出现异常，返回1。
```

#### 全量对⽐

采集间隔：1 day
报警手段：邮件+zabbix
采集规则：

```
步骤1 定时任务执行脚本，查询中央数据库，获取心跳时间(字段名message_timestamp)比当前时间差小于350s和状态(字段名push_state)为4的记录，记录包含cid；
步骤2；查询通配数据库，获取记录包含cid、视频周期、图片周期，并与'步骤1'中的cid对比，取cid交集的记录，过滤出白名单中的cid；
步骤3 根据'步骤2'得到的cid列列表，循环调用对象存储'checkserver'的接口，获取最新的视频周期和图片周期；
步骤4 根据'步骤2'和'步骤3'返回的结果进行比对，同一个cid在通配的视频(图片)的存储周期与对象存储的视频（图⽚）的存储周期不一致则认定异常结果，若'checkserver'接口返回的cid的time字段为0，cycle为-1，该种情况（cid从未推流或者上传图片）记录为异常周期cycle记为'-1'。异常结果记录对应的cid、通配的视频周期、通配的图片周期、对象存储的视频周期、对象存储的图片周期，形成一个表格，同时生成一个附件；
步骤5 根据'步骤4'的异常结果和附件，发送邮件通知相关人；同时记录日志，正常结果记录0，异常结果记录1，供zabbix采集。
步骤6 脚本任意环节出现异常，发送异常邮件通知程序开发的负责人。
```

### 白名单规则说明

```
all：全局cid白名单
picture：查询图片存储周期的cid白名单
video：查询视频存储周期的cid白名单
1、对比'picture','video'是否存在公有的cid，如果存在将公有的从各自的列表中剔除，如果不存在全局白名单，则添加进全局白名单；
2、过滤完全局白名单之后，在查询对应存储周期时在分别过滤对应的cid白名单；
3、查询后cid的信息补全，如果查询cid图片存储周期，则对应视频存储周期使用通配数据库中查询到的存储周期补全（本应是从对象存储接口获取），反之亦然。适用于cid只在其中一个白名单（除全局白名单外）。
```



### 配置说明

```yaml
limit:
  cid_num: 16
  message_timestamp: 350000 # 纳秒
  step: 2000 # 全量查询时每次传多少个cid
  timeout: 2 # 分钟
log:
  exfile: cid.xlsx
  filename: check_storage_cycle.log
  format: json
  level: info
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
    server: smtp.mxhichina.com # 公司邮箱配置smtp使用这个，不需要调整
    subject: 【存储周期每日全量检测】 南昌云
    username: xxxxx@topvdn.com
mongodb:
  camera:
    db: camera
    fields:
      - _id
      - message_timestamp
      - push_state
    table: device
    url: mongodb://camera:xxxxx@monghost:27017/?authSource=camera&authMechanism=SCRAM-SHA-1
  devices:
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
  mawarapp:
    db: mawarapp
    fields:
      - _id
      - group
    table: mawarapp_camera
    url: mongodb://mawarapp:xxxxx@monghost:27017/?authSource=mawarapp&authMechanism=SCRAM-SHA-1
url:
  checkServer: http://127.0.0.1:7061
  relay:
    password: ''
    url: http://127.0.0.1/stat
    username: ''
whitelist:
  all: null
  picture: null
  video: null
```

### 使用说明

#### 查看帮助信息

```
./check_storage_cycle -h
COMMANDS:
     full, f    全量对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致
     sample, s  抽样对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致
     help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

```

#### 抽样对比

```
./check_storage_cycle s
0
返回0正常，返回1有cid的视频或图片存储周期不匹配
```

#### 全量对比

```
./check_storage_cycle f
0
返回0正常，返回1有cid的视频或图片存储周期不匹配
```

