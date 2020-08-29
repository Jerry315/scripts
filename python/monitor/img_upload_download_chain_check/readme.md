# 对象存储图片上传下载链路健康检查

#### 使用说明

- 环境需要安装requests，PyYAML

- 启动参数共有四种（oss, oss_key，event, marwar）

  ```python
  # 方法一 模拟 upload2方式上传对象（不指定 key）
  python img_upload_download_chain_check.py -i oss
  # 方法二 模拟key方式的 upload2 上传对象
  python img_upload_download_chain_check.py -i oss_key
  # 方法三 模拟萤石上传抓拍事件
  python img_upload_download_chain_check.py -i event
  # 方法四 模拟爱耳目上传抓拍事件
  python img_upload_download_chain_check.py -i marwar
  ```

#### 配置说明

api_url：中央api地址

oss_url：对象存储系统接口

cid：测试cid

app_id：第三方app_id

app_key：第三方app_key

#### 逻辑说明

方法一：

1. 根据cid获取token，并返回token和cid信息；
2. 使用步骤1返回token和cid，调用/upload2，调用成功，返回obj_id，失败则重试3次，依旧失败，返回空{};
3. 如果步骤2失败，返回1，否则，使用步骤2中的obj_id和token，调用/file接口下载上传文件，比对下载后的文件大小和上传文件大小，一样大，则正常，否则抛出异常，返回1；
4. 以上步骤皆正常，返回0。

方法二：

1. 根据cid获取token，并返回token和cid信息；
2. 使用步骤1返回token和cid，用uuid随机生成一段字符串同固定开头的字符串"/1/"拼接起来作为key（"/1/d4c8295a-72f4-42be-a9d3-756530e52ea9"），携带key、token和cid调用/upload2，调用成功，返回obj_id和key，失败则重试3次，依旧失败，返回空{};
3. 如果步骤2失败，返回1，否则，使用步骤2中的token和key，调用/file接口下载上传文件，比对下载后的文件大小和上传文件大小，一样大，则正常，否则抛出异常，返回1；
4. 以上步骤皆正常，返回0。

方法三：

1. 根据cid获取token，并返回token和cid信息；

2. 使用步骤1返回token和cid，调用/upload3，调用成功，返回event url，失败则重试三次，依旧失败返回{};

3. 如果步骤2失败，返回1，否则使用步骤2中的event url和token，调用/files2接口下载上传文件，比对下载后的文件大小和上传文件大小，一样大，则正常，否则抛出异常，返回1；

4. 以上步骤皆正常，返回0。

   ```
   注：请求参数message和上传图片需要注意事项
    message = {
                   "topic_id": 0,
                   "channel_id": 0,
                   "subject": "",
                   "body": {},
                   "delay_time": 1000,
                   "attachments": [{
                       "form_field": "upload3", #要跟上传文件的名字对应
                       "key": "",
                       "area_id": 0,
                       "metadata": {},
                       "url": "",
                       "file_name": "",
                       "expiretype": 2 #周期要正确周期类型，0表示永久，1表示7天，2表示30天，3表示90天，默认为0
                   }]}
    f = {"upload3": ("upload3.jpg", open(upload3_img, 'rb'), 'image/jpg', {'Expires': '0'}),"message": json.dumps(message)} #此处的“upload3”需要跟message的“form_field”对应
   ```


方法4：

1. 根据cid获取token，并返回token和cid信息；
2. 使用步骤1返回的token，调用/iermu/uploadImg接口，记录上传时间戳，失败重试三次，依旧失败，返回空{}，成功返回token，上传时间戳和上传文件大小；
3. 使用步骤2返回的token、上传时间戳和上传文件大小，调用/fileinfo/last_objs?接口，根据结果返回的upload_time同上传时间戳比较相差在30s以内，获取obj_id，成功则返回token、obj_id以及上传文件大小，否则返回空{}；
4. 使用步骤3中的obj_id和token，调用/file接口下载上传文件，比对下载后的文件大小和上传文件大小，一样大，则正常，否则抛出异常，返回空{}；
5. 以上步骤皆正常，返回0，否则返回1。

注：目前跳过步骤四验证，如果步骤三正常，直接进入步骤5。