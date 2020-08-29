# -*- coding: utf-8 -*-
import time
import json
import optparse
import sys
from datetime import datetime, timedelta
from smtplib import SMTPServerDisconnected
from utils import stats
from utils import excel
from utils.common import ReportMail, AlertMail, Logger, zip_dir
from settings import report_file, config, cid_files, zip_name


def create_assist_date(start_time=None, end_time=None):
    # 创建日期列表
    datestart = datetime.fromtimestamp(start_time).strftime("%Y-%m-%d")
    dateend = datetime.fromtimestamp(end_time).strftime("%Y-%m-%d")

    # 转换为日期格式
    datestart = datetime.strptime(datestart, "%Y-%m-%d")
    dateend = datetime.strptime(dateend, "%Y-%m-%d")
    date_list = []
    while datestart <= dateend:
        tmp_date = datestart.strftime("%Y-%m-%d")
        date_list.append(tmp_date)
        datestart += timedelta(days=1)
    return date_list


def store_hour_data(cid_info_url, device_info_url, timeout=60):
    try:
        stats.gather_data(cid_info_url, device_info_url, timeout=timeout)
    except Exception as e:
        logger.error(str(e))
        alert_mail.send_plain(alert_receivers, subject, str(e))


def create_excel(period=7):
    t = time.time()
    end_time = time.mktime(time.strptime(
        datetime.fromtimestamp(t).replace(hour=0, minute=0, second=0).strftime(
            '%Y-%m-%d %H:%M:%S'), '%Y-%m-%d %H:%M:%S')) - 1
    start_time = (end_time - 86400 * period + 1)
    dateend = datetime.fromtimestamp(end_time).strftime("%Y-%m-%d")
    with open('%s.%s' % (cid_files, dateend), 'r') as c:
        cid_info = json.load(c)
    excel.write_excel(stats.get_period_data, cid_info, start_time, end_time, **{"is_bind": config["is_bind"]})
    zip_dir(report_file, zip_name)
    report_mail = ReportMail()
    report_mail.send_file(receivers, subject, zip_name)


if __name__ == '__main__':
    logger = Logger()
    start_time = int(time.time())
    receivers = config['smtp']['receivers']
    subject = config['smtp']['subject']
    alert_receivers = config['alert']['smtp']['receivers']
    alert_subject = config['alert']['smtp']['subject']
    cid_info_url = config['cid_info_url']
    device_info_url = config['device_info_url']
    timeout = config['timeout']
    print "start time ", start_time
    try:
        usage = 'python main.py [-s] [-p<period>]'
        parser = optparse.OptionParser(usage)
        parser.add_option('-p', '--period', dest='period', help='several days ')
        parser.add_option('-s', '--store', action="store_false", help='store data hourly ')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        if argv[0] == '-s':
            store_hour_data(cid_info_url, device_info_url, timeout=timeout)
        else:
            period = options.period
            try:
                if period:
                    period = int(period)
            except Exception:
                period = 7
            create_excel(period=period)
    except SMTPServerDisconnected as e:
        alert_mail = AlertMail()
        alert_mail.send_plain(alert_receivers, alert_subject, str(e))
    except Exception as e:
        logger.error(str(e))

    print "cost time %s" % (int(time.time()) - start_time)
