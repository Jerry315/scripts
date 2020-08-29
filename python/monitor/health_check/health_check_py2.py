#!/usr/bin/env python
# -*- coding: utf-8 -*-
import optparse, sys, requests, time, yaml, re, urllib2, json, urllib3
import logging, os
from logging.handlers import RotatingFileHandler
from bs4 import BeautifulSoup
from random import sample

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
config = yaml.load(open(os.path.join(BASE_DIR, 'config.yaml'), 'r'))
log_dir = config['log_dir']['log_dir']
if log_dir:
    RUN_LOG_FILE = os.path.join(log_dir, 'health_check_run.log')
    ERROR_LOG_FILE = os.path.join(log_dir, 'health_check_error.log')
    UNUSUAL_LOG_FILE = os.path.join(log_dir, 'health_check_unusual.log')
else:
    RUN_LOG_FILE = os.path.join(BASE_DIR, 'log/health_check_run.log')
    ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log/health_check_error.log')
    UNUSUAL_LOG_FILE = os.path.join(BASE_DIR, 'log/health_check_unusual.log')


class Logger(object):
    __instance = None

    def __init__(self):
        self.run_log_file = RUN_LOG_FILE
        self.error_log_file = ERROR_LOG_FILE
        self.unusual_log_file = UNUSUAL_LOG_FILE
        self.run_logger = None
        self.error_logger = None

    def __new__(cls, *args, **kwargs):
        if not cls.__instance:
            cls.__instance = object.__new__(cls, *args, **kwargs)
        return cls.__instance

    @staticmethod
    def check_path_exist(log_abs_file):
        log_path = os.path.split(log_abs_file)[0]
        if not os.path.exists(log_path):
            os.makedirs(log_path)

    def initialize_run_log(self, level):
        self.check_path_exist(self.run_log_file)
        file_1_1 = RotatingFileHandler(filename=self.run_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('run_log', level=level)
        logger1.addHandler(file_1_1)
        return logger1

    def initialize_error_log(self, level):
        self.check_path_exist(self.error_log_file)
        file_1_1 = RotatingFileHandler(filename=self.error_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s  - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('error_log', level=level)
        logger1.addHandler(file_1_1)
        return logger1

    def initialize_unusual_log(self):
        self.check_path_exist(self.run_log_file)
        file_1_1 = RotatingFileHandler(filename=self.unusual_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s : %(message)s")
        file_1_1.setFormatter(fmt=fmt)
        logger1 = logging.Logger('run_log', level=logging.WARNING)
        logger1.addHandler(file_1_1)
        return logger1

    def debug(self, msg):
        logger = self.initialize_run_log(logging.DEBUG)
        logger.debug(msg)

    def info(self, msg):
        logger = self.initialize_run_log(logging.INFO)
        logger.info(msg)

    def warn(self, msg):
        logger = self.initialize_unusual_log()
        logger.warn(msg)

    def error(self, msg):
        logger = self.initialize_error_log(logging.ERROR)
        logger.error(msg, exc_info=True)


class Check(object):
    def __init__(self, api_url, token_url, oss_url):
        self.api_url = api_url
        self.token_url = token_url
        self.oss_url = oss_url
        self.log = Logger()
        self.msg = {
            'cid': None,
            'msg': None,
            'code': None
        }

    def parse_xml(self, stat_url, auth_user=None, auth_pwd=None):
        '''
        从stat接口获取cid相关信息，随机取出2个cid对bw_in和time两个字段进行判断，是否满足需求（bw_in>100 && time > 60s）
        :param stat_url:
        :param auth_user:
        :param auth_pwd:
        :return:
        '''
        status_code = 0
        try:
            if auth_user and auth_pwd:
                auth_handler = urllib2.HTTPPasswordMgrWithDefaultRealm()
                auth_handler.add_password(None, stat_url, auth_user, auth_pwd)
                urllib2.install_opener(urllib2.build_opener(urllib2.HTTPBasicAuthHandler(auth_handler)))
                page = urllib2.urlopen(stat_url, timeout=5)
            else:
                page = urllib2.urlopen(stat_url, timeout=5)
            # 使用BeautifulSoup对返回的数据进行解析
            soup = BeautifulSoup(page, 'html5lib')
            # 获取所有stream元素
            streams = soup.find_all('stream')

            data = []
            for stream in streams:
                bw_in = int(stream.find('bw_in').text)
                cost_time = int(stream.find('time').text)
                if (bw_in / 1024 > 100) and (cost_time / 60000 > 1):
                    tmp = {}
                    tmp['cid'] = stream.find('name').text
                    tmp['bw_in'] = bw_in
                    tmp['time'] = cost_time
                    data.append(tmp)
                else:
                    illegal_str = 'cid: %s, bw_in: %s,time: %s' % (stream.find('name').text, bw_in, cost_time)
                    self.log.warn(illegal_str)
            if not data:
                raise Exception("has not cid match rule ")
        except Exception as e:
            self.msg['msg'] = str(e) + '[parse_xml: access stat_url error, bad url or request timeout or auth failed].'
            self.log.error(json.dumps(self.msg))
            return 1

        # 根据符合条件的stream的数量，在0到这个范围内随机生成两个数字
        ts = sample(range(0, len(data)), 2)
        cids = [int(data[num]['cid']) for num in ts]
        token_info = self.get_token(cids)
        if token_info:
            for token in token_info:
                status_code = 0
                if not self.get_record(token['cid'], token['token']):
                    if self.get_record(token['cid'], token['token']):
                        status_code = 0
                    else:
                        status_code = 1
        else:
            status_code = 1
        return status_code

    def get_record(self, cid, client_token):
        end = int(time.time())
        if self.oss_url.endswith('/'):
            self.oss_url.strip('/')
        url = self.oss_url + '/oss/v1/{cid}/record/timeline?begin={begin}&end={end}&client_token={client_token}'.format(
            **{'cid': cid, 'begin': end - 3600, 'end': end, 'client_token': client_token})
        try:
            headers = {
                'Content-Type': 'application/json'
            }
            result = requests.get(url, headers=headers, timeout=30)
            self.msg['code'] = result.status_code
            data = result.json()['timelines']
            if not data:
                raise Exception("has not timeline")
            differ = (end - data[-1]['end'])
            if differ < 60:
                if self.get_m3u8(cid, client_token, data[-1]['end'] - 60, data[-1]['end']):
                    return True
            else:
                self.msg['cid'] = cid
                self.msg[
                    'msg'] = 'get_record: The current time is more than one minute longer than the video record last end time .'
                self.log.error(self.msg)
        except Exception as e:
            self.msg['cid'] = cid
            self.msg['msg'] = str(e) + '[get_record: access url error,get video record failed].'
            self.log.error(self.msg)
        return False

    def get_m3u8(self, cid, token, start, end):
        if self.api_url.endswith('/'):
            self.api_url.strip('/')
        url = self.oss_url + "/oss/v1/{cid}/record/m3u8/{start}_{end}.m3u8?client_token={token}".format(**{
            "cid": cid,
            "start": start,
            "end": end,
            "token": token
        })
        headers = {
            'Content-Type': 'application/json'
        }
        try:
            result = requests.get(url, headers=headers, timeout=10)
            self.msg['code'] = result.status_code
            data = result.text
            if re.search(r'/record/ts', data):
                self.msg['cid'] = cid
                self.msg['msg'] = 'get_m3u8: get download_ts address success.'
                self.log.info(self.msg)
                return True
            else:
                self.msg['cid'] = cid
                self.msg['msg'] = 'get_m3u8: has not download_ts address found.'
                self.msg['status'] = False
                self.log.error(self.msg)
        except Exception as e:
            self.msg['cid'] = cid
            self.msg['msg'] = str(e) + '[get_m3u8: access url error, get download_ts address failed].'
            self.log.error(self.msg)
        return False

    def get_token(self, cids):
        token_list = []
        for cid in cids:
            url = self.token_url + str(cid)
            try:
                headers = {
                    'Content-Type': 'application/json'
                }
                response = requests.get(url, headers=headers, timeout=5)
                self.msg['code'] = response.status_code
                result = response.json()
                token_list.append({'cid': cid, 'token': result['token']})
                self.msg['cid'] = cid
                self.msg['msg'] = 'get token success'
                self.log.info(self.msg)
            except Exception as e:
                self.msg['cid'] = cid
                self.msg['msg'] = str(e) + '[get_token: access curl error,get token failed].'
                self.log.error(self.msg)
                return
        return token_list


if __name__ == '__main__':
    # python版本低于2.7.9以下需要加下面这一项，避免出现SNIMissingWarning告警
    urllib3.disable_warnings()
    username = config['httpbasicauth']['username']
    password = config['httpbasicauth']['password']
    stat_url = config['project_url']['stat_url']
    api_url = config['project_url']['api_url']
    oss_url = config['project_url']['oss_url']
    token_url = config['project_url']['tokern_url']
    parser = optparse.OptionParser()
    parser.add_option('-a', '--api', dest='api')
    parser.add_option('-s', '--stat', dest='stat')
    parser.add_option('-u', '--username', dest='username')
    parser.add_option('-p', '--pwd', dest='password')
    argv = sys.argv[1:]
    (options, args) = parser.parse_args(argv)
    if options.api:
        api_url = options.api
    if options.stat:
        stat_url = options.stat
    if options.username:
        username = options.username
    if options.password:
        password = options.password
    ck = Check(api_url, token_url, oss_url)
    ret = ck.parse_xml(stat_url, username, password)
    print(ret)
