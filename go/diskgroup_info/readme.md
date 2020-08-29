## 磁盘调度格式化

参数说明：

```
summary 查看总体概要或者详细信息，后面加上-d参数查看详细
	-d、--detail 是否查看详细，默认不传

group 查看磁盘具体信息
	-c、--cycle 磁盘组存储周期
	-g、--groupId 磁盘组id
	-d、--detail 是否查看详细，默认不传

host 主机视图查看具体信息
	-c、--cycle 磁盘组存储周期
	-g、--groupId 磁盘组id
	-H、--host 主机ip
```

规则说明：

```
默认磁盘类型为2
0号组，过滤超过一天未上传心跳的磁盘
```

使用说明：

```
获取所有磁盘组概要信息
./diskgroup_info summary
[2018-06-20 11:41:50] [Getdiskinfo; group_id: 1, group_type: 2, cycle: 30, onlines: 12, 
[2019-04-16 17:56:52] [cycle_30: 6, cycle_7: 8, cycle_0: 1]
获取所有磁盘组详细信息
./diskgroup_info summary -d
[2018-06-20 11:41:50] [Getdiskinfo; group_id: 1, group_type: 2, cycle: 30, onlines: 12, lockTime: 0, dispatcher_id: [], total_used: 15.01TB, total_left: 25.72TB]
	[first_time: 2019-02-25 09:47:56, seq: 0, discID: 538968136, is_online: 1, public_ip: 113.105.173.91, local_ip: 192.168.2.91, port: 9106, left: 2.16TB]
	[first_time: 2019-02-25 09:47:56, seq: 1, discID: 538968185, is_online: 1, public_ip: 113.105.173.92, local_ip: 192.168.2.92, port: 9110, left: 2.16TB]

获取某一磁盘组的概要信息
./diskgroup_info group -g 9
[2018-09-28 20:03:13] [Getdiskinfo; group_id: 9, group_type: 2, cycle: 7, onlines: 10, lockTime: 0, dispatcher_id: [436207624], total_used: 13.76TB, total_left: 20.18TB]

获取某一磁盘组的详细信息
./diskgroup_info group -g 9 -d
[2018-09-28 20:03:13] [Getdiskinfo; group_id: 9, group_type: 2, cycle: 7, onlines: 10, lockTime: 0, dispatcher_id: [436207624], total_used: 13.76TB, total_left: 20.18TB]

获取存储周期为7的磁盘组概要信息
./diskgroup_info group -c 7
[2018-09-28 20:03:13] [Getdiskinfo; group_id: 9, group_type: 2, cycle: 7, onlines: 10, lockTime: 0, dispatcher_id: [436207624], total_used: 13.76TB, total_left: 20.18TB]

获取存储周期为7的磁盘组详细信息
./diskgroup_info group -c 7 -d
[2018-09-28 20:03:13] [Getdiskinfo; group_id: 9, group_type: 2, cycle: 7, onlines: 10, lockTime: 0, dispatcher_id: [436207624], total_used: 13.76TB, total_left: 20.18TB]

获取存储周期为7，磁盘组id为9的磁盘组详细信息
./diskgroup_info group -c 7 -g 9 -d
[2018-09-28 20:03:13] [Getdiskinfo; group_id: 9, group_type: 2, cycle: 7, onlines: 10, lockTime: 0, dispatcher_id: [436207624], total_used: 13.76TB, total_left: 20.18TB]

主机视图查看磁盘组信息
./diskgroup_info host
+--------------+---------+------+-------+-----------+--------+
|   LOCALIP    | GROUPID | PORT | CYCLE |  DISCID   | STATUS |
+--------------+---------+------+-------+-----------+--------+
| 192.168.2.89 |       1 | 9101 |    30 | 539885600 | Y      |
+              +---------+------+       +-----------+        +
|              |      11 | 9102 |       | 539885609 |        |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9103 |     0 | 539885618 | N      |
+              +---------+------+-------+-----------+--------+
|              |       6 | 9104 |    15 | 539885627 | Y      |
+              +---------+------+-------+-----------+        +
|              |       9 | 9105 |     7 | 539885636 |        |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9106 |     0 | 539885645 | N      |
+              +---------+------+-------+-----------+--------+
|              |       5 | 9107 |    15 | 539885654 | Y      |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9108 |     0 | 539885663 | N      |
+              +---------+------+-------+-----------+--------+
|              |       7 | 9109 |    30 | 539885672 | Y      |
+              +---------+------+       +-----------+        +
|              |       8 | 9110 |       | 539885681 |        |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9111 |     0 | 539885690 | N      |
+              +---------+------+-------+-----------+--------+
|              |       1 | 9112 |    30 | 539885698 | Y      |
+              +---------+------+-------+-----------+        +
|              |      10 | 9113 |    15 | 539885706 |        |
+--------------+---------+------+-------+-----------+--------+

根据主机ip查看磁盘组信息
./diskgroup_info host -H 192.168.2.87
+--------------+---------+------+-------+-----------+--------+
|   LOCALIP    | GROUPID | PORT | CYCLE |  DISCID   | STATUS |
+--------------+---------+------+-------+-----------+--------+
| 192.168.2.87 |       8 | 9101 |    30 | 539885581 | Y      |
+              +---------+------+-------+-----------+        +
|              |       9 | 9102 |     7 | 539885582 |        |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9103 |     0 | 539885583 | N      |
+              +         +------+       +-----------+        +
|              |         | 9104 |       | 539885584 |        |
+              +---------+------+-------+-----------+--------+
|              |       4 | 9105 |     7 | 539885585 | Y      |
+              +---------+------+-------+-----------+        +
|              |      10 | 9106 |    15 | 539885586 |        |
+              +---------+------+       +-----------+        +
|              |       6 | 9107 |       | 539885587 |        |
+              +---------+------+-------+-----------+        +
|              |       3 | 9108 |     7 | 539885588 |        |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9109 |     0 | 539885589 | N      |
+              +---------+------+-------+-----------+--------+
|              |       5 | 9110 |    15 | 539885590 | Y      |
+              +---------+------+-------+-----------+--------+
|              |       0 | 9111 |     0 | 539885591 | N      |
+              +---------+------+-------+-----------+--------+
|              |       7 | 9112 |    30 | 539885592 | Y      |
+              +---------+------+       +-----------+        +
|              |      11 | 9113 |       | 539885593 |        |
+--------------+---------+------+-------+-----------+--------+

根据存储周期查看磁盘组信息

 ./diskgroup_info host -c 30
+--------------+---------+------+-------+-----------+--------+
|   LOCALIP    | GROUPID | PORT | CYCLE |  DISCID   | STATUS |
+--------------+---------+------+-------+-----------+--------+
| 192.168.2.95 |       1 | 9103 |    30 | 539885621 | Y      |
+              +---------+------+       +-----------+        +
|              |       7 | 9104 |       | 539885630 |        |
+              +---------+------+       +-----------+        +
|              |      11 | 9106 |       | 539885648 |        |
+              +---------+------+       +-----------+        +
|              |       8 | 9108 |       | 539885666 |        |
+--------------+         +------+       +-----------+        +

根据groupId查看磁盘组信息
 ./diskgroup_info host -g 9
+--------------+---------+------+-------+-----------+--------+
|   LOCALIP    | GROUPID | PORT | CYCLE |  DISCID   | STATUS |
+--------------+---------+------+-------+-----------+--------+
| 192.168.2.93 |       9 | 9110 |     7 | 539885683 | Y      |
+--------------+         +      +       +-----------+        +
| 192.168.2.88 |         |      |       | 539885733 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.90 |         | 9108 |       | 539885670 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.87 |         | 9102 |       | 539885582 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.89 |         | 9105 |       | 539885636 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.98 |         | 9111 |       | 539885693 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.94 |         | 9106 |       | 539885644 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.95 |         | 9105 |       | 539885639 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.96 |         | 9111 |       | 539885686 |        |
+--------------+         +------+       +-----------+        +
| 192.168.2.92 |         | 9108 |       | 539885637 |        |
+--------------+---------+------+-------+-----------+--------+


```
