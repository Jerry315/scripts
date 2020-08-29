#!/usr/bin/env python
# -*- coding: utf-8 -*-
'''
运行环境：python2.7，需要安装requests模块
在运行之前可以传入要查看的group_id，只显示指定group信息
url = 'http://codisktracker.private:8201/console/group_info'
params = {'disk_type':2}
'''
import optparse
import sys
import time
import requests
import json
from logger import Logger
from settings import config


class OSS(object):
    def __init__(self):
        self.logger = logger

    def parse_data(self, url, disk_type):
        result = requests.get(url, params={"disk_type": disk_type}, timeout=15)
        data = result.json()['group_infos']
        d = int(time.time())
        for item in data:
            msg = {}
            if disk_type == 3:
                msg['cycle'] = 99999
            else:
                msg['cycle'] = item.get('cycle')
            msg['time'] = d
            msg['group_id'] = item.get('group_id')
            msg['onlines'] = item.get('onlines')

            dispatcher_ids = item.get('dispatcher_ids')
            StorageScheme = item.get("StorageScheme")
            if not dispatcher_ids:
                dispatcher_id = None
            else:
                if isinstance(dispatcher_ids, list):
                    dispatcher_id = dispatcher_ids[0]
                else:
                    dispatcher_id = dispatcher_ids
            msg['dispatcher_id'] = dispatcher_id
            msg["StorageScheme"] = str(StorageScheme[0]) + "+" + str(StorageScheme[1])
            disc_infos = item.get('disc_infos')
            total_capacity = 0
            total_used = 0
            total_count = 0
            for disk in disc_infos:
                total_count += 1
                total_capacity += disk.get('left') + disk.get('used')
                total_used += disk.get('used')
            msg['total_capacity'] = total_capacity
            msg['total_used'] = total_used
            msg['total_count'] = total_count
            self.logger.log(json.dumps(msg))


if __name__ == '__main__':
    logger = Logger(config['oss']['codisktracker']['access_log'], config['oss']['codisktracker']['error_log'])
    try:
        parser = optparse.OptionParser()
        parser.add_option('-t', '--disk_type', dest='disk_type')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        url = config['oss']['codisktracker']['url']
        disk_type = options.disk_type
        try:
            disk_type = int(disk_type)
        except Exception:
            disk_type = 2
        oss = OSS()
        oss.parse_data(url, disk_type)
    except Exception as e:
        logger.log(str(e), mode=False)
