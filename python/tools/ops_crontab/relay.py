#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json
import time
import requests
from logger import Logger
from settings import config


class Relay(object):
    def __init__(self):
        self.logger = logger

    def get_data(self, relay_url):
        data = {}
        request = requests.get(relay_url, timeout=15).json()
        data['relay_name'] = request['relay_name']
        data['time'] = int(time.time())
        data['rtmp_bw_in'] = request['rtmp_bw_in']
        data['rtmp_bw_out'] = request['rtmp_bw_out']
        data['rtmp_connections_in'] = request['rtmp_connections_in']
        data['rtmp_connections_out'] = request['rtmp_connections_out']
        return data

    def parse_data(self, relay_url):
        data = self.get_data(relay_url)
        self.logger.log(json.dumps(data))


if __name__ == '__main__':
    logger = Logger(config['relay']['access_log'], config['relay']['error_log'])
    try:
        relay_url = config['relay']['url']
        Relay().parse_data(relay_url)
    except Exception as e:
        logger.log(str(e), mode=False)
