## 统计摄像机在线率

### agent

#### 说明

```
agent为server提供接口，从通配和中央数据取出数据
```

#### 配置

```yaml
config:
  mongodb:
    devices:
      url: mongodb://devices:xxxxx@monghost:27017/?authSource=devices&authMechanism=SCRAM-SHA-1
      db: devices
      table: devices_camera
      fields:
        "_id": 1
        "sn": 1
        "name": 1
        "brand": 1
        "model": 1
    mawarapp:
      db: mawarapp
      fields:
        "_id": 1
        "group": 1
        "is_bind": 1
      table: mawarapp_camera
      url: mongodb://mawarapp:xxxxx@monghost:27017/?authSource=mawarapp&authMechanism=SCRAM-SHA-1
    camera:
      url: mongodb://camera:xxxxx@monghost:27017/?authSource=camera&authMechanism=SCRAM-SHA-1
      db: camera
      table: device
      fields:
        "_id": 1
        "message_timestamp": 1
        "push_state": 1
  host: 127.0.0.1
  port: 9595
```



#### 部署

```
环境初始化
	./install.sh
启动agent
	将camera_stats_agent.conf文件放置/etc/supervisor/projects/
	执行：supervisorctl -c /etc/supervisor/supervisord.conf update
```

### server

#### 规则

```
1、定时采集统计
    1.1. 每小时查询中央数据库(表名：device)的数据，记为device_info，查询通配数据库(表名：mawar_camera)，存入camera_info：
        1.1.1、device_info数据包含的字段：_id、message_timestamp、push_state
        1.1.2、camera_info数据包含的字段：_id、sn、name、group、is_bind、brand、model
    1.2 将 device_info 数据分成两部分：
        1.2.1、取message_timestamp的值与当前时间比较，差值小于350s且push_state为4的cid标记为在线设备，push_state不为4的标记为离线设备，在线状态标记为 'Y'，离线状态标记为'N'
        1.2.2、取message_timestamp的值与当前时间比较，差值大于350s的cid标记为离线设备，并将在线状态标记为 'N'
    1.3、过滤 camera_info 数据，将 camera_info 存在的 cid 但不存在于 device_info 的 cid 对应记录剔除掉
    1.4、数据入库：
        1.4.1、camera_info 数据写入本地文件(发邮件用到)
        1.4.2、device_info 数据更新写入数据库
2、定时发送统计结果
    发送时间：每周五上午9：15
    统计周期：默认一周(上周五00:000:00到本周四23:59:59)

```

#### 配置

```bash
config:
  mongodb:
    url: 'mongodb://camera:xxxxxxxxxxxx@127.0.0.1:27017/?authSource=camera&authMechanism=SCRAM-SHA-1'
    db: 'camera'
  cid_info_url: 'http://127.0.0.1:5000/camera/stats/v1/cid_info'
  device_info_url: 'http://127.0.0.1:5000/camera/stats/v1/device_info'
  timeout: 60000
  step: 1000
  pool_size: 100
  username: "jerry" # 获取token所用的用户
  password: xxxxx # 获取token用户的密码
  smtp:
    subject: 【数据统计】摄像机在线率统计报表
    user: 'xxxxxxxxx@qq.com'
    passwd: 'xxxxxxx'
    smtp_server: 'smtp.qq.com'
    receivers:
      - 'zhanlin@antelope.cloud'
  alert: # 数据插入异常告警邮箱
    smtp:
      subject: 【数据统计】摄像机在线率统计报表--异常
      user: 'xxxxxx@qq.com'
      passwd: 'xxxxxxxxxx'
      smtp_server: 'smtp.qq.com'
      receivers:
        - 'zhanlin@antelope.cloud'
```

#### 部署

```
环境初始化
	./install.sh
启动api
	将camera_stats_server.conf文件放置/etc/supervisor/projects/
	执行：supervisorctl -c /etc/supervisor/supervisord.conf update
```

#### 每小时获取数据的计划任务

```bash
30 */1 * * root /bin/bash /opt/camera_stats/server/start.sh -s >/dev/null 2>&1 &
```

#### 周报发送计划任务

```
*/30 * * * 5 root /bin/bash /opt/camera_stats/server/start.sh -p 7 >/dev/null 2>&1 &
```

#### 接口使用说明

##### 获取token

接口：/camera/v1/token

方法：post

参数

```json
{
	"username": username,
	"password": password
}
```
请求样例

```
curl -X POST \
      http://127.0.0.1:5001/camera/v1/token \
      -H 'Content-Type: application/json' \
      -d '{
        "username": "jerry",
        "password": 123456
    }'
```

返回数据

```
	response = {
        "code": 0,
        "msg": "登录成功",
        "token": xxxxxxxxxxx
    }
```

状态码说明

| 状态码 | 解释           |
| ------ | -------------- |
| 0      | 正常           |
| 40001  | 用户名或密码错 |

##### 请求周期数据

接口：/camera/v1/records

方法：get

参数

| 字段  | 类型   | 是否必传 | 说明        |
| ----- | ------ | -------- | ----------- |
| cids  | array  | 否       | 指定cid列表 |
| end   | int    | 是       | 结束时间戳  |
| start | int    | 是       | 开始时间戳  |
| page  | int    | 否       | 不传默认为1 |
| size  | int    | 否       | 不传默认100 |
| token | string | 是       | 鉴权需要    |

请求样例

```
curl -X GET \
  'http://127.0.0.1:5000/camera/v1/records?token=xxxx&start=1562428800&end=1562436000&cids=[538443798,20538443824]&size=10&page=1' 
```

返回数据

```json
response = {
        "code": 0,
        "msg": "获取数据成功",
        "data": [
            {
                "cid": 538378294,
                "create_time": 1562430002,
                "group": "ungrouped",
                "name": "解绑失败",
                "sn": "137898508619",
                "status": "N"
            }
        ]
    }
```

状态码说明

| 状态码 | 说明              |
| ------ | ----------------- |
| 0      | 正常              |
| 40001  | token失效         |
| 40002  | cid 列表超过10000 |
| 40003  | 查询时间轴超过7天 |
| 40004  | 参数不合法        |
| 500    | 程序出错          |

