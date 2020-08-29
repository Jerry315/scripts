## 上传文件到对象存储

- 配置说明

  ```yaml
  appId: "xxxx" 
  appKey: "xxxxxxxx"
  apiUrl: "https://dgdx-api.antelopecloud.cn/v2/devices/tokens"
  ossUrl: "https://dgdx-oss1.antelopecloud.cn"
  cid: 539099876
  expiretype: 1 #存储周期
  upload:
    path: "/opt/backup"
    layout: "20060102" # 重要，备份文件中包含的日期格式，如果格式不匹配对应的文件则不能备份
  download:
    path: "/opt/download"
    objIds: # 下载文件的objectid
      - xxxxxxxxxxxxxx 
      - xxxxxxxxxxxxxx
  dataName: "backup.txt" # 保存上传和下载文件的信息
  log:
    level: info
    path: ""
    filename: "backup_oss.log"
  ```

- 使用说明

  ```
  ./backuToOss upload 上传文件，根据配置文件中，upload.layout时间格式获取执行脚本时datetime字符串，例如当前日期是2019年4月26日，则会匹配上传目录下，文件名包含有20190426的所有文件。
  ./backuToOss download 下载文件，根据配置文件中download.objIds去下载文件
  ```

  

