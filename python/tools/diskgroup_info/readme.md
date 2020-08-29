## 磁盘调度格式化

参数说明：

```
-g: 磁盘组id
-p: 磁盘组存储周期，0，7，30
-t: disk_type，磁盘类型 1：循环存储磁盘 2：循环对象磁盘 3：永久对象磁盘
-u: 磁盘调度接口
-d: 显示各磁盘组下详细信息，默认不显示磁盘组详细信息
```

规则说明：

```
0号组，过滤超过一天未上传心跳的磁盘
```

使用说明：

```
获取所有磁盘类型为2的磁盘组信息
python diskgroup_info.py -t 2

获取某一磁盘组的概要信息
python diskgroup_info.py -t 2 -g 23

获取某一磁盘组的详细信息
python diskgroup_info.py -t 2 -g 23 -d

获取存储周期为7的磁盘组信息
python diskgroup_info.py -t 2 -p 7
```

