## 磁盘组速率和容量监控

参数说明：

- -m或--method，要调用的方法；
- -g或--gid，磁盘组id；
- -p或--percent，显示磁盘剩余容量百分比。

监控规则：

- 磁盘组剩余容量监控规则
  - 通过-m group方法返回group_id，磁盘组周期平均上传速度，磁盘剩余空间预警阀值（根据磁盘大小决定，为磁盘容量的二分之一，如4TB磁盘，预警阀值为2TB）；
  - 查询0号组容量，过滤掉心跳时间超过1天，状态不在线的；
  - 剩余磁盘容量百分比小于20%同时剩余空间小于预警阀值，触发预警。
- 磁盘组速率监控规则
  - 通过-m group方法返回group_id，磁盘组周期平均上传速度，磁盘剩余空间预警阀值（根据磁盘大小决定，为磁盘容量的二分之一，如4TB磁盘，预警阀值为2TB）；
  - 磁盘组5分钟的平均速率，如果连续两次大于磁盘组周期平均速率，触发告警。

使用说明：

1. monitor_diskgroup_capacity

   ```
   查询返回磁盘组的id列表
   ./monitor_diskgroup_capacity.sh normal -m group
   {"data": [{"{#GROUPID}": "2", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "3", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "4", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "5", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "6", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "7", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "8", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "9", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "10", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "11", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "12", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "15", "{#MAX_UPLOAD_RATE}": "1439917", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "16", "{#MAX_UPLOAD_RATE}": "1439917", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "19", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "20", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "22", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "23", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "24", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "25", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "26", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "27", "{#MAX_UPLOAD_RATE}": "2843294", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "28", "{#MAX_UPLOAD_RATE}": "2843294", "{#LIMIT_CAPACITY}": "4"}]}
   参数说明：
   	GROUPID：磁盘组id
   	MAX_UPLOAD_RATE：磁盘组周期内的平均速率
   	LIMIT_CAPACITY：磁盘组预警阀值
   
   根据磁盘组id，返回指定组剩余容量
   ./monitor_diskgroup_capacity.sh normal -m capacity -g 24
   8628647743488
   
   ./monitor_diskgroup_capacity.sh normal -m capacity -g 24 -p true
   97.19
   
   0号组容量查询
   ./monitor_diskgroup_capacity.sh zero
   {'data': [{'{#DISK_SPACE}': '4', '{#STORAGE_CYCLE}': '0'}, {'{#DISK_SPACE}': '4', '{#STORAGE_CYCLE}': '30'}, {'{#DISK_SPACE}': '4', '{#STORAGE_CYCLE}': '7'}]}
   参数说明：
   	DISK_SPACE：磁盘类型4TB或者8TB
   	STORAGE_CYCLE：存储周期
   获取周期为7，磁盘类型为4TB磁盘的剩余容量
   ./monitor_diskgroup_capacity.sh zero -c 7 -s 4 -a left
   179146479016960
   获取周期为7，磁盘类型为4TB磁盘的数量
   ./monitor_diskgroup_capacity.sh zero -c 7 -s 4 -a count
   48
   ```

2. monitor_diskgroup_rate

   ```
   查询返回磁盘组的id列表
   python monitor_diskgroup_capacity.py -m group
   {"data": [{"{#GROUPID}": "2", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "3", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "4", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "5", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "6", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "7", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "8", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "9", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "10", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "11", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "12", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "15", "{#MAX_UPLOAD_RATE}": "1439917", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "16", "{#MAX_UPLOAD_RATE}": "1439917", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "19", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "20", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "22", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "23", "{#MAX_UPLOAD_RATE}": "6171076", "{#LIMIT_CAPACITY}": "2"}, {"{#GROUPID}": "24", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "25", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "26", "{#MAX_UPLOAD_RATE}": "5686589", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "27", "{#MAX_UPLOAD_RATE}": "2843294", "{#LIMIT_CAPACITY}": "4"}, {"{#GROUPID}": "28", "{#MAX_UPLOAD_RATE}": "2843294", "{#LIMIT_CAPACITY}": "4"}]}
   
   根据磁盘组id，返回指定组5分钟内平均上传速率
   python monitor_diskgroup_rate.py -m rate -g 2
   420727
   ```