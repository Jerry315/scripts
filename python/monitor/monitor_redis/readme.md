redis监控

1、环境初始化

```bash
./install.sh
```

2、更新配置config.yml

```yaml
config:
  redis:
    password: cc
    db: 0 # 默认db
```

3、使用说明

```
# 获取帮助
 ./monitor_redis.sh -h

# 获取主机上运行的redis信息
 ./monitor_redis.sh -q
# 获取cpu信息
 ./monitor_redis.sh ip port -c

# 获取内存信息
./monitor_redis.sh ip port -m

# 获取db0的信息
 ./monitor_redis.sh ip port -d 0

```

