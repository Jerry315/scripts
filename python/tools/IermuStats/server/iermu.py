# -*- coding: utf-8 -*-
import requests
import os
import datetime
from prettytable import PrettyTable
from common import config, Logger, SendMail, RUN_LOG_FILE


def iermu_stat(limit):
    result = {}
    for url in config['stat_url']:
        # area_count = 0
        # repeat_count = 0
        data = requests.get(url,params={"limit": limit})
        if data.status_code == 200:
            for cid, value in data.json().items():
                if cid == 'date':
                    continue
                if int(cid) in result:
                    # repeat_count += 1
                    result[int(cid)] = result[int(cid)] + int(value)
                else:
                    result[int(cid)] = int(value)
        #         area_count += 1
        # print 'area_count: ', area_count
        # print 'repeat_count: ', repeat_count
    result = sorted(result.items(), key=lambda x: x[1], reverse=True)
    tb = PrettyTable(["CID", "Value"])
    count = 0
    for item in result:
        tb.add_row([item[0], item[1]])
        count += 1
    logger.log(str(tb))
    text = '统计日期：%s\n统计内容：统计报警事件条数超过%d条的cid的数量，共有%d条记录。' % (
        (datetime.date.today() - datetime.timedelta(days=1)).strftime('%Y%m%d'),
        limit, count)
    SendMail().send_file(receivers, subject, text, RUN_LOG_FILE)


if __name__ == '__main__':
    if os.path.exists(RUN_LOG_FILE):
        os.remove(RUN_LOG_FILE)
    logger = Logger()
    try:
        receivers = config['smtp']['receivers']
        subject = config['smtp']['subject']
        limit = config['limit']
        iermu_stat(limit)
    except Exception as e:
        logger.log(str(e),mode=False)
