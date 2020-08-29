ops_crontab说明

1、产生mawar日志

```bash
*/5 * * * * root /bin/bash /opt/sa_tools/ops_crontab/start.sh mawar
```

2、生成codistracker日志

```
*/5 * * * * root /bin/bash /opt/sa_tools/ops_crontab/start.sh codistracker
```

3、生成relay日志

```
*/5 * * * * root /bin/bash /opt/sa_tools/ops_crontab/start.sh relay
```

