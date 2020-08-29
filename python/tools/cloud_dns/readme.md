## 云解析域名操作脚本

1. 配置说明settings.py

   ```python
   ###############aliyun arm info###############
   AK = "xxxxx" # 阿里云AccessKeyID
   SECRET = "xxxx" # 阿里云AccessKeySecret
   REGION_ID = "default" # 默认
   ENDPOINT = "alidns.aliyuncs.com" # 默认
   FORMAT = "JSON" # 也支持xml
   VERSION = "2015-01-09" # 根据官方
   PROTOCOL = "https" # http或https
   METHOD = "POST"
   
   ##############log file#######################
   RUN_LOG_FILE = os.path.join(BASE_DIR, 'logs', 'cloud_dns.access.log')
   ERROR_LOG_FILE = os.path.join(BASE_DIR, 'logs', 'cloud_dns.error.log')
   
   ##############records file 批量操作模板文件###################
   RECORDS_FILE = os.path.join(BASE_DIR,'records.xlsx')
   ```

2. 环境初始化

   ```
   ./install.sh
   ```

3. 批量操作使用records.xlsx模板文件

4. 使用说明

   ```
   查看脚本帮助
    ./ddns.sh --help
   Usage: ddns.py [OPTIONS] COMMAND [ARGS]...
   
   Options:
     --help  Show this message and exit.
   
   Commands:
     add-domain      添加一条解析记录 domain: 根域名，baidu.com rr: 主机名，www type_: 解析类					  A记录、NS记录、MX记录等；value: 记录值，A记录对应ip
     batch-add-domain     从模板文件中导入域名记录，批量添加
     batch-delete-domain  从模板文件中导入域名记录，批量删除
     batch-set-domain     从模板文件中导入域名记录，批量禁用或启用
     batch-update-domain  从模板文件中导入域名记录，批量更新
     del-domain      删除解析记录 rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时					会返回
     del-sub-domain  删除主机记录对应的解析记录 domain: 域名名称 rr: 主机记录,www type_: 如果不填				   应的全部解析记录类型，解析类型包括(不区分大小写)：A、MX、CNAME、TXT、						  DIRECT_URL、FORWORD_URL、NS、AAAA、SRV
     get-all-domain  获取当前账户下所有域名的信息
     get-one-domain  根据传入参数获取指定主域名的所有解析记录列表 domain: 根域名 pg: 如果返回页数较				  多，显示那一页 ps: 每一页显示多少条记录
     set-domain      设置解析记录状态 rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列					表时会返回 status:Enable启用解析 Disable: 暂停解析
     update-domain   更新一条解析记录 rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列					表时会返回 rr:主机记录,www；type_: 如果不填写，对应的全部解析记录类型。 解析类					 型包括(不区分大小写)：A、MX、CNAME、TXT、REDIRECT_URL、FORWORD_URL、NS、AAAA、SRV
   
   查看当前账号下管理域名信息
   ./ddns.sh get-all-domain
   {
       "RequestId": "F641158D-0315-4DBD-B5B5-693AE640194B",
       "Domains": {
           "Domain": [
               {
                   "RecordCount": 2,
                   "AliDomain": true,
                   "VersionCode": "mianfei",
                   "DomainId": "5b27b443-1624-4463-a0db-31c2773b8320",
                   "VersionName": "Alibaba Cloud DNS",
                   "PunyCode": "xxx.com",
                   "DomainName": "xxx.com",
                   "DnsServers": {
                       "DnsServer": [
                           "dns13.hichina.com",
                           "dns14.hichina.com"
                       ]
                   }
               }
           ]
       },
       "TotalCount": 1,
       "PageNumber": 1,
       "PageSize": 20
   }
   
   获取某个域名下解析条目
   ./ddns.sh get-one-domain xxx.com
   
   {
       "PageNumber": 1,
       "TotalCount": 2,
       "RequestId": "5499DCE3-D8EA-4F4F-B401-06F570B734A3",
       "DomainRecords": {
           "Record": [
               {
                   "Type": "A",
                   "DomainName": "xxx.com",
                   "RR": "ai",
                   "Line": "default",
                   "TTL": 600,
                   "Weight": 1,
                   "RecordId": "17231412996474880",
                   "Status": "ENABLE",
                   "Value": "123.5.56.78",
                   "Locked": false
               },
               {
                   "Type": "A",
                   "DomainName": "xxx.com",
                   "RR": "test",
                   "Line": "default",
                   "TTL": 600,
                   "Weight": 1,
                   "RecordId": "17226828302189568",
                   "Status": "DISABLE",
                   "Value": "10.10.10.11",
                   "Locked": false
               }
           ]
       },
       "PageSize": 20
   }
   
   
   新增一条解析
   ./ddns.sh add-domain xxx.com t1 A 192.168.2.55
   {
       "RequestId": "B61F6F37-3C14-4606-83E2-E7FF2EA79FDF",
       "RecordId": "17231680413333504"
   }
   
   设置解析启用或者禁用
   启用某个禁用的域名
   ./ddns.sh set-domain 17226828302189568 Enable
   {
       "RecordId": "17226828302189568",
       "RequestId": "D77F0AC1-B84E-464A-A163-BC4F2E65FAF9",
       "Status": "Enable"
   }
   
   更新一条解析
   ./ddns.sh update-domain 17226828302189568 t2 A 192.168.2.56
   {
       "RequestId": "D3A90D7A-C048-46DB-8B9F-F68B974AE814",
       "RecordId": "17226828302189568"
   }
   
   
   删除一条解析
   ./ddns.sh del-domain RecordId(17226828302189568)
   {
       "RequestId": "31ED2C25-B318-461A-8A8C-D3A06F6A69F0",
       "RecordId": "17226828302189568"
   }
   
   删除主机记录对应的解析记录
   ./ddns.sh del-sub-domain xxx.com t3 A
   {
       "TotalCount": "1",
       "RR": "t3",
       "RequestId": "777051EA-0ECB-4576-8D90-89D142019FCC"
   }
   
   批量添加域名
   ./ddns.sh batch-add-domain
   [
       {
           "Type": "A",
           "msg": "add record success",
           "Line": "default",
           "DomainName": "xxx.com",
           "RecordId": "17237239023864832",
           "RR": "zl",
           "Value": "52.13.14.99"
       },
       {
           "Type": "A",
           "msg": "add record success",
           "Line": "unicom",
           "DomainName": "xxx.com",
           "RecordId": "17237239060169728",
           "RR": "ll",
           "Value": "52.13.14.99"
       },
       {
           "Type": "A",
           "msg": "add record success",
           "Line": "telecom",
           "DomainName": "xxx.com",
           "RecordId": "17237239097329664",
           "RR": "gg",
           "Value": "52.13.14.99"
       }
   ]
   
   批量删除域名
   ./ddns.sh batch-delete-domain
   [
       {
       	"msg": "delete record success",
           "RecordId": "17237172859438080"
       },
       {
       	"msg": "delete record success",
           "RecordId": "17237172898237440"
       },
       {
       	"msg": "delete record success",
           "RecordId": "17237172931987456"
       }
   ]
   批量更新域名
   ./ddns.sh batch-add-domain
   [
       {
           "Type": "A",
           "msg": "update record success",
           "Line": "default",
           "RecordId": "17237239023864832",
           "RR": "zl",
           "Value": "52.13.14.99"
       },
   ]
   
   批量禁用或者启用域名
   ./ddns.sh batch-set-domain
   [
       {
           "Status": "Disable",
           "RecordId": "17237239097329664",
           "msg": "Disable record success"
       }
   ]
   
   
   
   ```

   

