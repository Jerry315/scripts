#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json
import optparse
import sys
import time
import urllib2
import gevent
import requests
import pandas as pd
from datetime import datetime
from bs4 import BeautifulSoup
from prettytable import PrettyTable
from common import Logger, SendMail, config


class CheckCidTimeOut:
    def __init__(self, auth_user=None, auth_pwd=None):
        self.auth_user = auth_user
        self.auth_pwd = auth_pwd
        self.logger = logger
        self.relay_urls = config['relay']['urls']
        self.response = {}
        self.data = []
        self.cids = []
        self.__storate__ = dict()

    def parse_cids(self, stat_url):
        try:
            if self.auth_user and self.auth_pwd:
                auth_handler = urllib2.HTTPPasswordMgrWithDefaultRealm()
                auth_handler.add_password(None, stat_url, self.auth_user, self.auth_pwd)
                urllib2.install_opener(urllib2.build_opener(urllib2.HTTPBasicAuthHandler(auth_handler)))
                page = urllib2.urlopen(stat_url, timeout=5)
            else:
                page = urllib2.urlopen(stat_url, timeout=5)
            # 使用BeautifulSoup对返回的数据进行解析
            soup = BeautifulSoup(page, 'html5lib')
            # 获取所有stream元素
            streams = soup.find_all('stream')
        except Exception:
            self.logger.log('parse_cids: access %s failed.' % stat_url, mode=False)
            return
        for stream in streams:
            bw_in = int(stream.find('bw_in').text)
            cost_time = int(stream.find('time').text)
            if (bw_in / 1024 > 100) and (cost_time / 60000 > 1):
                cid = int(stream.find('name').text)
                if cid in config['white_list']:
                    continue
                self.cids.append(int(stream.find('name').text))
                if stat_url not in self.__storate__:
                    self.__storate__[stat_url] = []
                self.__storate__[stat_url].append(int(stream.find('name').text))

    def get_cids(self):
        gevent.joinall([gevent.spawn(self.parse_cids, stat_url) for stat_url in self.relay_urls])
        self.cids = list(set(self.cids))

    def check(self, checkmode):
        web_url = config['checkserver']['url']
        if config['checkserver']['timeout']:
            timeout = config['checkserver']['timeout']
        else:
            timeout = 120
        if config['step']:
            step = config['step']
        else:
            step = 100
        headers = {
            'content-type': "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
        }
        querystring = {"timeout": timeout, "checkmode": checkmode}
        cid_len = len(self.cids)
        index = 0
        while True:
            period_cid = self.cids[index:index + step]
            payload = "------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"cidlist\"\r\n\r\n{\"cidlist\": %s}\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--" % period_cid
            result = requests.post(web_url, data=payload, headers=headers, params=querystring)
            status_code = result.status_code
            if status_code == 200:
                result = result.json()
                if result.get('timeoutcids'):
                    self.data += result.get('timeoutcids')
            else:
                self.logger.log(
                    'check: checkserver url: %s access checkserver failed, checkmode: %s, status_code: %s' % (
                        web_url, checkmode, status_code),
                    mode=False)
            if (cid_len - index) < step and (cid_len - index) > 0:
                index += (cid_len - index)
            elif cid_len - index > step:
                index += step
            else:
                break
            time.sleep(10)

    def agg_data(self, checkmode):
        ignore_urls = config['checkmode'][checkmode]['ignore_url']
        if ignore_urls:
            for ignore_url in ignore_urls:
                if ignore_url in self.relay_urls:
                    self.relay_urls.remove(ignore_url)
        self.get_cids()
        self.check(checkmode)
        if not self.data:
            return
        df = pd.DataFrame(self.data)
        if checkmode == 0:
            df = df.sort_values(by='LatestImageTime')
            df['LatestImageTime'] = df.apply(
                lambda x: datetime.fromtimestamp(x['LatestImageTime']).strftime('%Y-%m-%d %H:%M:%S'), axis=1)
            df['LatestVideoTime'] = df.apply(
                lambda x: datetime.fromtimestamp(x['LatestVideoTime']).strftime('%Y-%m-%d %H:%M:%S'), axis=1)
            check_data = df.to_dict('records')

            t = PrettyTable(['CID', 'LatestImageTime', 'LatestVideoTime'])
            for item in check_data:
                t.add_row([item['CID'], item['LatestImageTime'], item['LatestVideoTime']])
        else:
            df = df.sort_values(by='LatestVideoTime')
            df['LatestVideoTime'] = df.apply(
                lambda x: datetime.fromtimestamp(x['LatestVideoTime']).strftime('%Y-%m-%d %H:%M:%S'), axis=1)
            check_data = df.to_dict('records')

            t = PrettyTable(['CID', 'LatestVideoTime'])
            for item in check_data:
                t.add_row([item['CID'], item['LatestVideoTime']])
                for k,v in self.__storate__.items():
                    if item['CID'] in v:
                        logger.warn(json.dumps({"unusual_cid": item['CID'],"relay_url": k, "checkmode": checkmode}))
        self.logger.warn(str(t))
        self.response['check_cid_num'] = len(self.cids)
        self.response['unusual_cid_num'] = len(check_data)
        return str(t)


if __name__ == '__main__':
    logger = Logger()
    parser = optparse.OptionParser()
    parser.add_option('-m', '--checkmode', dest='checkmode', help='checkmode is 0 or 1')
    argv = sys.argv[1:]
    (options, args) = parser.parse_args(argv)
    checkmode = options.checkmode
    try:
        checkmode = int(checkmode)
    except Exception:
        checkmode = 0
    subject = config['checkmode'][checkmode]['subject'] + '({})'.format(datetime.today().strftime("%Y-%m-%d"))
    try:
        auth_user = config['auth']['auth_user']
        auth_pwd = config['auth']['auth_pwd']
        receivers = config['smtp']['receivers']
        start_time = time.time()
        cct = CheckCidTimeOut(auth_user, auth_pwd)
        msg = cct.agg_data(checkmode)
        end_time = time.time()
        cct.response['cost time'] = int(end_time - start_time)
        if not msg:
            cct.response['check_cid_num'] = len(cct.cids)
            cct.response['unusual_cid_num'] = 0
            cct.response['msg'] = 'no unusual cid'
            logger.info(json.dumps(cct.response))
        else:
            logger.warn(json.dumps(cct.response))
            msg = json.dumps(cct.response) + '\n' + msg
            SendMail().send(receivers, subject, msg)
    except Exception as e:
        SendMail().send(['zhanlin@antelope.cloud'], subject, 'Something cause an error, %s' % str(e))
        logger.log(str(e), mode=False)
