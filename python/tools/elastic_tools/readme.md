## elasticsearch索引操作脚本

1. 脚本配置说明

   ```
   elastic:
     timeout: 15
     url: http://192.168.2.82:9200
     max_snapshot_bytes_per_sec: '50mb'
     max_restore_bytes_per_sec:  '100mb'
     indexs: # 索引部分
       snapshot: # 需要备份的索引
         - index:
           -
             elastic-test-
           enable: 1 # enable为1时操作该条目所有索引
           delay_days: 1
           date_fmt: "y.m.d" # 时间格式，例如：ymd，y.m.d, y-m-d
         - index:
           -
           enable: 0 # enable为0，操作时忽略该条目所有索引
           delay_days:
           date_fmt: "y.m.d"
       delete: # 需要删除的索引或者备份
         - index:
           -
             elastic-test-
           enable: 1
           delay_days: 1
           date_fmt: "y.m.d"
         - index:
           -
           enable: 0
           delay_days:
           date_fmt: "y.m.d"
       setting: # 需要设置标签的索引
         - index:
           -
             elastic-test-
           enable: 1
           delay_days: 1
           date_fmt: "y.m.d"
           tag: test
         - index:
           -
           enable: 0
           delay_days:
           date_fmt: "y.m.d"
           tag:
   ```

2. 初始化脚本环境

   ```
   执行
   ./install.sh
   ```

3. 使用说明

   ```
   获取帮助
   ./es_handle.sh --help
   Usage: es_handle.py [OPTIONS] COMMAND [ARGS]...
   
   Options:
     --help  Show this message and exit.
   
   Commands:
     del-snapshot  删除指定日期备份
     delete        删除配置文件中指定的索引
     get-snapshot  获取备份索引信息
     put-tag       给索引设置标签
     snapshot      备份配置文件中指定的索引
     
   备份索引
   ./es_handle.sh snapshot
   
   删除索引
   ./es_handle.sh delete
   
   获取指定日期备份索引信息
   ./es_handle.sh get-snapshot
   {
       "snapshots": [
           {
               "uuid": "N9-qI0WaT7aqDDzEoKaXXQ", 
               "duration_in_millis": 1932, 
               "start_time": "2019-02-28T03:13:31.227Z", 
               "shards": {
                   "successful": 5, 
                   "failed": 0, 
                   "total": 5
               }, 
               "version_id": 5061599, 
               "end_time_in_millis": 1551323613159, 
               "state": "SUCCESS", 
               "version": "5.6.15", 
               "snapshot": "snapshot", 
               "end_time": "2019-02-28T03:13:33.159Z", 
               "indices": [
                   "elastic-test-2019.02.27"
               ], 
               "failures": [], 
               "start_time_in_millis": 1551323611227
           }
       ]
   }
   
   
   删除指定日期备份索引
   ./es_handle.sh del-snapshot
   
   给索引设置标签
   ./es_handle.sh put-tag
   
   ```

   