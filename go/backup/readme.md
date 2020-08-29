## 数据备份脚本

- 配置说明

  ```yaml
  layout: "20060102" 
  expire: 30
  files:
    - /showdoc_data/html/Sqlite/showdoc.db.php
  dirs:
    -
      dir: /showdoc_data/html/Public/Uploads
      cmd: 
        - "tar -zcvf Uploads-`date +%Y%m%d.tar.gz` -C /showdoc_data/html/Public Uploads"
    -
      dir: /opt/atlassian/data
      cmd:
        - "tar --exclude confluence/backups --exclude jira/export -zcvf atlassian-`date +%Y%m%d`.tar.gz -C /opt/atlassian/ data"
        - "split -b 200M -d -a 1 atlassian-`date +%Y%m%d`.tar.gz atlassian-`date +%Y%m%d`.tar.gz"
        - "rm  atlassian-`date +%Y%m%d`.tar.gz"
  dbs:
    -
      host: 192.168.2.210
      port: 3306
      username: jira
      password: jira
      database: jira
    -
      host: 192.168.2.210
      port: 3306
      username: confluence
      password: confluence
      database: confluence
  backup: /public/backup
  
  # 关于目录备份的命令，这里提前定义好，方便过滤文件和目录。
  ```

- 功能说明

  ```
  1、备份文件，备份配置文件中指定的文件到backup目录，添加当天日期后缀
  2、备份目录，打包配置文件中指定的目录到backup目录，添加当天日期后缀
  3、备份数据库，备份配置文件中指定的数据库到backup目录，添加当天日期后缀
  4、所有备份动作完成后，清理超过30天的备份数据
  ./backup 自动执行上述动作
  ```

  

