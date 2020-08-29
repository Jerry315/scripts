#!/usr/bin/env python
# -*- coding: utf-8 -*-
### 通配后台数据收集 ###
import json
import time
import requests
from logger import Logger
from settings import config


class Mawar(object):
    def __init__(self):
        self.logger = logger
        self.msg1 = {"category": "all", "data": {}}
        self.msg2 = {'category': "brand", "data": {}}
        self.msg3 = {"category": "model", "data": {}}
        self.msg4 = {"category": "storage", "data": {}}

    def parse_data(self, url):
        stat_url = url + "/tong/v1/ops/stat/onoffline-num"
        storage_url = url + "/tong/v1/ops/stat/storage-num"
        n = int(time.time())
        stat_data = requests.get(stat_url, timeout=30).json()['data']
        storage_data = requests.get(storage_url, timeout=30).json()["data"]
        temp = dict()
        for key, value in storage_data.items():
            temp[key] = dict()
            for k, v in value.items():
                if k == "":
                    continue
                temp[key][k] = v
        self.msg4["data"] = temp
        self.msg1['time'] = n
        self.msg1['data'] = {
            "total_num": stat_data['total_num'],
            "online_num": stat_data['online_num'],
            "offline_num": stat_data['offline_num']
        }
        self.msg2['time'] = n
        self.msg3['time'] = n
        self.msg4['time'] = n
        for brand in stat_data['brand']:
            self.msg3['data'][brand] = {}
            self.msg2['data'].update({brand: {"online_num": stat_data['brand'][brand]['online_num'],
                                              "offline_num": stat_data['brand'][brand]['offline_num']}})
            for mod in stat_data['brand'][brand]['model']:
                self.msg3['data'][brand].update(
                    {mod: {"online_num": stat_data['brand'][brand]['model'][mod]['online_num'],
                           "offline_num": stat_data['brand'][brand]['model'][mod]['offline_num']
                           }})
        self.logger.log(json.dumps(self.msg1).decode('utf-8'))
        self.logger.log(json.dumps(self.msg2).decode('utf-8'))
        self.logger.log(json.dumps(self.msg3).decode('utf-8'))
        self.logger.log(json.dumps(self.msg4).decode("utf-8"))


class DetailMawar(object):
    def __init__(self):
        self.logger = logger
        self.msg2 = {'category': "brand", "data": {}}
        self.msg3 = {"category": "model", "data": {}}
        self.msg4 = {"category": "storage", "data": {}}

    def parse_data(self, url):
        stat_url = url + "/tong/v1/ops/stat/onoffline-num"
        storage_url = url + "/tong/v1/ops/stat/storage-num"
        n = int(time.time())
        stat_data = requests.get(stat_url, timeout=30).json()['data']
        storage_data = requests.get(storage_url, timeout=30).json()["data"]
        for key, value in storage_data.items():
            for k, v in value.items():
                if k == "":
                    continue
                for b, c in v["brand"].items():
                    self.logger.log(
                        json.dumps(
                            {'category': "brand", "media": key, "cycle": int(k), "brand": b, "count": c, "time": n}))
                self.logger.log(json.dumps(
                    {'category': "media_cycle", "media": key, "cycle": int(k), "count": v["total"], "time": n,
                     "datasource": "mawar"}))
        for brand in stat_data['brand']:
            for m, d in stat_data['brand'][brand]["model"].items():
                self.logger.log(
                    json.dumps({"category": "model", "brand": brand, "model": m, "offline_num": d["offline_num"],
                                "online_num": d["online_num"], "time": n}))


if __name__ == '__main__':
    logger = Logger(config['mawar']['access_log'], config['mawar']['error_log'])
    try:
        mawar_url = config['mawar']['url']
        Mawar().parse_data(mawar_url)
        DetailMawar().parse_data(mawar_url)
    except Exception as e:
        logger.log(str(e), mode=False)
