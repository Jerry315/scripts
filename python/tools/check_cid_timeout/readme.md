# 检测CID超时

#### 使用说明

- 环境需要安装bs4 requests pyyaml html5lib prettytable pandas gevent；
- 启动参数 -m checkmode（checkmode：0，检测图片上传；1，检测推流）。

#### 配置说明

- relay.urls：所有转发的地址，角色为A0的忽略；
- auth：配置转发的httpbasic认证；
- webserver.url:  webserver地址，该接口需要使用内网ip访问；
- webserver.timeout：设置检查cid的超时时间；
- step：每次查询cid个数。
- checkmode：根据checkmode选择邮件主题，ignore_urls忽略检查的relay_url

#### 规则说明

1. 推流大小(bw_in > 100kb)并且推流时间(time > 1分钟)；
2. cid超时时间12小时；

#### 逻辑说明

1. 通过转发地址，获取所有cid，所有cid满足规则说明中的规则1，cid去重；
2. 根据step大小，每次取step个cid进行检查，将返回结果放在列表中，直到所有cid都检查完毕，统计出所有cid数量，以及满足规则说明中的规则2的cid数量，结束循环；
3. 根据checkmode，如果是0对步骤2中所有的cid根据LatestImageTime进行从小到大排序，如果是1对步骤2中所有的cid根据LatestVideoTime进行从小到大排序；
4. 对步骤3排序后的数据写入到日志文件；
5. 异常情况邮件发送指定收件人。

### 定时任务
0  9 * * *    root bash /opt/sa_tools/check_cid_timeout/check_cid_timeout.sh -m 0
30  9 * * *    root bash /opt/sa_tools/check_cid_timeout/check_cid_timeout.sh -m 1