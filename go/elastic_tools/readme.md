## elasticsearch索引操作脚本

1. 脚本配置说明

   ```
   timeout: 15
   esUrl: "http://192.168.2.82:9200"
   maxSnapshotBytesPerSec: "50mb"
   maxRestoreBytesPerSec: "100mb"
   location: "opt/es_backup/" # 快照备份目录
   snapshot:
     - index:
         - elastic-test-
       enable: true
       delayDays: 1
       dateFmt: "20060102" # 根据生成索引格式来定
   delete:
     - index:
         - elastic-test-
       enable: false
       delayDays: 1
       dateFmt: "20060102"
   settings:
     - index:
         - elastic-test-
       enable: false
       delayDays: 1
       dateFmt: "20060102"
       tag:
   log:
     level: info
     Path: ""
     Filename: "elastic_tools.log"
   ```
   
   
   
3. 使用说明

   ```
   获取帮助
   ./elastic_tools -h
   NAME:
      operate elastic indices - A new cli application
   
   USAGE:
      elastic_tools [global options] command [command options] [arguments...]
   
   VERSION:
      0.0.0
   
   COMMANDS:
        delete        delete indices # 删除指定索引
        set-tag       set indices tag # 为索引设置tag
        repository    create snapshot repository # 创建索引快照仓库位置
        snapshot      create indices snapshot # 创建索引快照
        get-snapshot  get indices snapshot # 获取指定仓库的快照
        del-snapshot  delete snapshot # 删除指定仓库的快照
        get-tag       get indices tag # 获取指定索引的tag
        help, h       Shows a list of commands or help for one command
   
   GLOBAL OPTIONS:
      --help, -h     show help
      --version, -v  print the version
   
     
   备份索引
   ./elastic_tools snapshot
   
   删除索引
   ./elastic_tools delete
   
   获取指定仓库快照索引信息
   ./elastic_tools get-snapshot -r 20190524
   {
       "snapshots": [
           {
               "snapshot": "20190524",
               "uuid": "ISCFUifFTLWSQW-furLYqg",
               "version_id": 5061599,
               "version": "5.6.15",
               "indices": [
                   "elastic-test-2019.05.23"
               ],
               "state": "SUCCESS",
               "reason": "",
               "start_time": "2019-05-24T02:36:55.39Z",
               "start_time_in_millis": 1558665415390,
               "end_time": "2019-05-24T02:36:57.244Z",
               "end_time_in_millis": 1558665417244,
               "duration_in_millis": 1854,
               "failures": [
                   
               ],
               "shards": {
                   "total": 5,
                "successful": 5,
                   "failed": 0
               }
           }
       ]
   }
   
   
   删除指定仓库下某快照
   ./elastic_tools del-snapshot -r my_backup -s snapshot
   
   给索引设置标签
   ./elastic_tools set-tag
   
   获取某索引的标签
   /elastic_tools get-tag -i elastic-test-2019.05.23
   {
       "elastic-test-2019.05.23": {
           "settings": {
               "index": {
                   "creation_date": "1558579629310",
                   "number_of_replicas": "1",
                   "number_of_shards": "5",
                   "provided_name": "elastic-test-2019.05.23",
                   "refresh_interval": "-1",
                   "routing": {
                       "allocation": {
                           "require": {
                               "tag": "haha"
                           }
                       }
                   },
                   "uuid": "2-uX3pbZSV-g8Nw4bawxig",
                   "version": {
                       "created": "5061599"
                   }
               }
           }
       }
   }
   
   ```
   
   