#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
import time
import datetime

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
config_file = os.path.join(BASE_DIR, 'conf', 'config.yml')
config = yaml.load(open(config_file, 'rb'))

if config.get("log_dir", None) is None:
    RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'sendwechat.access.log')
    ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'sendwechat.error.log')
else:
    log_dir = config.get("log_dir")
    RUN_LOG_FILE = os.path.join(log_dir, 'sendwechat.access.log')
    ERROR_LOG_FILE = os.path.join(log_dir, 'sendwechat.error.log')

# 未恢复告警的记录文件
record_file_dir = config.get("record_file_dir")

# 微信企业信息，企业id，应用secret等信息
wechat_conf = config['wechat']
wechat_corpid = wechat_conf['corpid']
token_url = wechat_conf['token_url']
send_url = wechat_conf['send_url']
# 告警信息
alert_info = config['alert']
# redis的信息
redis_conf = config['redis']
redis_host = redis_conf['host']
redis_passwd = redis_conf['passwd']
redis_port = redis_conf['port']
redis_db = redis_conf['db']

# smtp信息
smtp_conf = config['smtp']
smtp_subject = smtp_conf['subject']
smtp_user = smtp_conf['user']
smtp_passwd = smtp_conf['passwd']
smtp_server = smtp_conf['smtp_server']
smtp_receivers = smtp_conf['receivers']

# 告警级别
Information = 5
Warning = 4
Average = 3
High = 2
Disaster = 1

# 时间格式
CURR_DATE_YEAR = time.strftime("%Y")
CURR_DATE_MONTH = time.strftime("%m")
CURR_DATE_DAY = time.strftime("%d")
CURR_DATE_MONTH_2 = time.strftime("%b")
CURR_DATE_2 = time.strftime("%Y-%m-%d")
CURR_TIME_4 = time.strftime("%H:%M:%S")
TODAY = datetime.date.today()
ONEDAY = datetime.timedelta(days=1)
TWODAY = datetime.timedelta(days=2)
THREEDAY = datetime.timedelta(days=3)
YESTERDAY = TODAY - ONEDAY
YESTERDAY_OF_YESTERDAY = TODAY - TWODAY
YESTERDAY_OF_YESTERDAY_OF_YESTERDAY = TODAY - THREEDAY
a = int(time.strftime("%w"))
b = int(time.strftime("%w")) + 6
delay_day_of_last_week_last = datetime.timedelta(days=a)
delay_day_of_last_week_first = datetime.timedelta(days=b)
NOW_HOUR = int(time.strftime("%H"))

time_format = {
    "CURR_DATE_YEAR": CURR_DATE_YEAR,
    "CURR_DATE_MONTH": CURR_DATE_MONTH,
    "CURR_DATE_DAY": CURR_DATE_DAY,
    "CURR_DATE_MONTH_2": CURR_DATE_MONTH_2,
    "CURR_DATE_1": time.strftime("%Y_%m_%d"),
    "CURR_DATE_2": CURR_DATE_2,
    "CURR_DATE_3": time.strftime("%Y%m%d"),
    "CURR_DATE_4": time.strftime("%Y.%m.%d"),
    "CURR_TIME_1": time.strftime("%H_%M_%S"),
    "CURR_TIME_2": time.strftime("%H-%M-%S"),
    "CURR_TIME_3": time.strftime("%H%M%S"),
    "CURR_TIME_4": CURR_TIME_4,
    "NOW_TIME": "%s %s" % (CURR_DATE_2, CURR_TIME_4),
    "NOW_TIME_2": "%s %s %s" % (CURR_DATE_MONTH_2, CURR_DATE_DAY, CURR_TIME_4),
    "TODAY": TODAY,
    "ONEDAY": ONEDAY,
    "TWODAY": TWODAY,
    "THREEDAY": THREEDAY,
    "YESTERDAY": YESTERDAY,
    "YESTERDAY_OF_YESTERDAY": YESTERDAY_OF_YESTERDAY,
    "YESTERDAY_OF_YESTERDAY_OF_YESTERDAY": YESTERDAY_OF_YESTERDAY_OF_YESTERDAY,
    "TOMORROW": TODAY + ONEDAY,
    "a": a,
    "b": b,
    "delay_day_of_last_week_last": delay_day_of_last_week_last,
    "delay_day_of_last_week_first": delay_day_of_last_week_first,
    "LAST_WEEK_LAST_DAY": TODAY - delay_day_of_last_week_last,
    "LAST_WEEK_FIRST_DAY": TODAY - delay_day_of_last_week_first,
    "FIRST_DAY_OF_THIS_MONTH": "%s-%s-1" % (CURR_DATE_YEAR, CURR_DATE_MONTH),
    "LAST_DAY_OF_THIS_MONTH": "%s-%s-31" % (CURR_DATE_YEAR, CURR_DATE_MONTH),
    "BEGIN_OF_TODAY": "%s 00:00:00" % CURR_DATE_2,
    "END_OF_TODAY": "%s 23:59:59" % CURR_DATE_2,
    "BEGIN_OF_YESTERDAY": "%s 00:00:00" % YESTERDAY,
    "END_OF_YESTERDAY": "%s 23:59:59" % YESTERDAY,
    "BEGIN_OF_YESTERDAY_OF_YESTERDAY": "%s 00:00:00" % YESTERDAY_OF_YESTERDAY,
    "END_OF_YESTERDAY_OF_YESTERDAY": "%s 23:59:59" % YESTERDAY_OF_YESTERDAY,
    "BEGIN_OF_YESTERDAY_OF_YESTERDAY_OF_YESTERDAY": "%s 00:00:00" % YESTERDAY_OF_YESTERDAY_OF_YESTERDAY,
    "END_OF_YESTERDAY_OF_YESTERDAY_OF_YESTERDAY": "%s 23:59:59" % YESTERDAY_OF_YESTERDAY_OF_YESTERDAY,
    "NOW_HOUR": NOW_HOUR,
    "start_date_of_hour_stat": "%s %s:00:00" % (CURR_DATE_2, NOW_HOUR),
    "end_date_of_hour_stat": "%s %s:59:59" % (CURR_DATE_2, NOW_HOUR)
}
#####################################################


# 告警分析地址
alert_url = config['alert_url']
