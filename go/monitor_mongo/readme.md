## mongo监控

#### 配置config.yml

```yaml
instance:
  - url: "mongodb://report:aFsuzppGDNHLTZt7CGyhkQGyz@192.168.2.72:27017/?authSource=report&authMechanism=SCRAM-SHA-1"
    db: "report"
    port: 27017
log:
  level: "info"
  path: ""
  fileName: "monitor_mongo.log"
  
 注：
 	这里用到的mongo用户名需要有mongodb的集群管理权限
 	timeout任务执行超过时间
```

#### 使用说明

- 获取监控mongo信息

  ```
  ./monitor_mongo query
  {"data": [{"{#PORT}": "27417", "{#CONF_FILE}": "/etc/mongodb/co_adfs_primary.conf"}, {"{#PORT}": "27517", "{#CONF_FILE}": "/etc/mongodb/co_disktracker_primary.conf"}]}
  ```

- 获取mongodb启动时间

  ```bash
  ./monitor_mongo uptime -p 27017
  24386454
  ```

- 获取进程信息

  ```bash
  ./monitor_mongo pid -c /etc/mongodb/mo_psrnc_secondary.conf
  1
  
  进程存在返回1，不存在返回0
  ```

- 获取mongod连接信息

  ```bash
  创建连接总数
  ./monitor_mongo connections -p 27017 -i TotalCreated
  可用连接数
  ./monitor_mongo connections -p 27017 -i Available
  4194
  当前连接数
  ./monitor_mongo connections -p 27017 -i Current
  2359
  ```

- 获取内存信息

  ```bash
  物理内存消耗
  ./monitor_mongo memory -p 27017 -i Resident
  虚拟内存消耗
  ./monitor_mongo memory -p 27017 -i Virtual
  映射内存消耗
  ./monitor_mongo memory -p 27017 -i Mapped
  
  MappedWithJournal：除了映射内存外还包括journal日志消耗的映射内存
  Mapped： 映射内存
  Resident： 物理内存消耗，单位M
  Bits： 操作系统位数
  Virtual： 虚拟内存消耗
  
  ```

- 获取网络信息（zabbix中增量形式显示）

  ```bash
  进入流量
  ./monitor_mongo network -p 27017 -i BytesIn   
  28840072
  
  进出流量
  ./monitor_mongo network -p 27017 -i BytesOut    
  28840072
  
  接收到不同请求的总数
  ./monitor_mongo network -p 27017 -i NumRequests   
  28840072
  
  ```

- 获取集群状态信息

  ```bash
  ./monitor_mongo.sh replhealth 27017
  1
  ```

- 获取复制延迟

  ```bash
  ./monitor_mongo.sh replrelay -p 27417
  1.0
  ```

- 获取当前任务数量

  ```bash
  ./monitor_mongo.sh op 27017
  4
  
  检测当前执行中的任务执行时间是否超过预设timeout时间，超过的任务记录到日志中。
  ```

- 获取数据库从启动后各种操作总共的数量（zabbix中增量形式显示）

  ```bash
  最后一次启动后的insert次数
   ./monitor_mongo opcounters -p 27017 -i Insert
  2186494
  
  最后一次启动后的query次数
   ./monitor_mongo opcounters -p 27017 -i Query
  210978630
  
  最后一次启动后的update次数
   ./monitor_mongo opcounters -p 27017 -i Update
  131199734
  
  最近一次启动后delete次数
   ./monitor_mongo opcounters -p 27017 -i Delete
  136450
  
  最后一次启动后的getmore次数
  ./monitor_mongo opcounters -p 27017 -i Getmore
  56747621
  
  最后一次启动后的command次数
  ./monitor_mongo opcounters -p 27017 -i Command
  1047980968
  ```

- 获取数据库**副本**从启动后各种操作总共的数量（zabbix中增量形式显示）

  ```bash
  最后一次启动后的insert次数
   ./monitor_mongo opcountersRepl -p 27017 -i Insert
  2186494
  
  最后一次启动后的query次数
   ./monitor_mongo opcountersRepl -p 27017 -i Query
  210978630
  
  最后一次启动后的update次数
   ./monitor_mongo opcountersRepl -p 27017 -i Update
  131199734
  
  最近一次启动后delete次数
   ./monitor_mongo opcountersRepl -p 27017 -i Delete
  136450
  
  最后一次启动后的getmore次数
  ./monitor_mongo opcountersRepl -p 27017 -i Getmore
  56747621
  
  最后一次启动后的command次数
  ./monitor_mongo opcountersRepl -p 27017 -i Command
  1047980968
  ```

- 获取全局锁相关信息

  ```bash
  当前的全局锁等待锁等待的个数
  ./monitor_mongo currentQueue -p 27017 -i Total
  0
  当前全局写锁等待个数
  ./monitor_mongo currentQueue -p 27017 -i Writers
  0
  当前的全局读锁等待个数
  ./monitor_mongo currentQueue -p 27017 -i Readers
  0
  
  当前实例活跃客户端数量
  ./monitor_mongo activeClients -p 27017 -i Total
  2384
  活跃客户端中写操作个数
  ./monitor_mongo activeClients -p 27017 -i Writers
  0
  活跃客户端读操作个数
  ./monitor_mongo activeClients -p 27017 -i Readers
  0
  ```

- 额外信息

  ```
  数据库访问数据时发现数据不在内存时的页面数量，当数据库性能很差或者数据量极大时，这个值会显著上升
  ./monitor_mongo extra_info -p 27017 -i Page_faults
  32951
  
  堆内存空间占用的字节数，仅linux适用
  ./monitor_mongoextra_info -p 27017 -i Heap_usage_bytes
  1526646592
  ```

  


