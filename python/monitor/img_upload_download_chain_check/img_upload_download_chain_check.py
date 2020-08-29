#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json
import logging
import optparse
import os
import sys
import time
from logging.handlers import RotatingFileHandler

import requests
import yaml
from requests.packages.urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'chain_check_run.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'chain_check_error.log')
small_01_img = os.path.join(BASE_DIR, 'img', 'small_01.jpg')
small_02_img = os.path.join(BASE_DIR, 'img', 'small_02.jpg')
upload3_img = os.path.join(BASE_DIR, 'img', 'upload3.jpg')
marwar_img = os.path.join(BASE_DIR, 'img', 'marwar.jpg')


class Logger(object):
    __instance = None

    def __init__(self):
        self.run_log_file = RUN_LOG_FILE
        self.error_log_file = ERROR_LOG_FILE
        self.run_logger = None
        self.error_logger = None

        self.initialize_run_log()
        self.initialize_error_log()

    def __new__(cls, *args, **kwargs):
        if not cls.__instance:
            cls.__instance = object.__new__(cls, *args, **kwargs)
        return cls.__instance

    @staticmethod
    def check_path_exist(log_abs_file):
        log_path = os.path.split(log_abs_file)[0]
        if not os.path.exists(log_path):
            os.makedirs(log_path)

    def initialize_run_log(self):
        self.check_path_exist(self.run_log_file)
        file_1_1 = RotatingFileHandler(filename=self.run_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s : %(message)s")
        file_1_1.setFormatter(fmt=fmt)
        logger1 = logging.Logger('run_log', level=logging.INFO)
        logger1.addHandler(file_1_1)
        self.run_logger = logger1

    def initialize_error_log(self):
        self.check_path_exist(self.error_log_file)
        file_1_1 = RotatingFileHandler(filename=self.error_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s : %(message)s")
        file_1_1.setFormatter(fmt=fmt)
        logger1 = logging.Logger('run_log', level=logging.ERROR)
        logger1.addHandler(file_1_1)
        self.error_logger = logger1

    def log(self, message, mode=True):
        if mode:
            self.run_logger.info(message)
        else:
            self.error_logger.error(message)


class ChainCheck(object):
    def __init__(self, api_url, oss_url):
        self.api_url = api_url
        self.oss_url = oss_url
        self.log = Logger()
        self.msg = {
            "code": None,
            "msg": None,
        }

    def get_token(self, cid, t=2):
        if self.api_url.endswith('/'):
            self.api_url.rstrip('/')
        url = self.api_url + '/v2/devices/tokens'
        token_info = {}
        try:
            headers = {
                'Content-Type': 'application/json',
                'X-App-Id': config['app_id'],
                'X-App-Key': config['app_key']
            }
            data = {
                "cids": cid
            }
            response = requests.post(url, headers=headers, json=data, verify=False, timeout=10)
            status_code = response.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                result = response.json()
                token_info.update(result['tokens'][0])
                self.msg['cid'] = cid
                self.msg['msg'] = 'get token success'
                self.log.log(self.msg)
            else:
                raise Exception('request code is not 200 ,get token failed')
        except Exception as e:
            self.msg['cid'] = cid
            self.msg['msg'] = 'get_token: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                token_info = self.get_token(cid, t)
                if token_info:
                    break
        return token_info

    def upload3(self, token_info, t=2):
        response = {}
        try:
            token = token_info['token']
            if self.oss_url.endswith('/'):
                self.oss_url.rstrip('/')
            url = self.oss_url + '/upload3?access_token={token}'.format(**{'token': token})
            message = {
                "topic_id": 0,
                "channel_id": 0,
                "subject": "",
                "body": {},
                "delay_time": 1000,
                "attachments": [{
                    "form_field": "upload3",
                    "key": "",
                    "area_id": 0,
                    "metadata": {},
                    "url": "",
                    "file_name": "",
                    "expiretype": expiretype
                }]}

            f = {"upload3": ("upload3.jpg", open(upload3_img, 'rb'), 'image/jpg', {'Expires': '0'}),
                 "message": json.dumps(message)}
            result = requests.post(url, files=f, verify=False, timeout=30)
            status_code = result.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                data = result.json()
                response['event_url'] = data['attachments'][0]['url']
                response['token'] = token
                response['up_size'] = os.path.getsize(upload3_img)
                self.msg['msg'] = 'upload3: upload img success'
                self.log.log(self.msg)
            else:
                raise Exception('request code is not 200, upload img failed')
        except Exception as e:
            self.msg['msg'] = 'upload3: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                response = self.upload3(token_info, t)
                if response:
                    break
        return response

    def download3(self, response, t=2):
        flag = 1
        # url = response['event_url'] + '?access_token=' + response['token']
        # url = response['event_url']
        if response["event_url"].__contains__("access_token"):
            url = response['event_url']
        else:
            url = response['event_url'] + '?access_token=' + response['token']
        try:
            data = requests.get(url, verify=False, timeout=15)
            status_code = data.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                down_size = len(data.content)
                self.msg['down_size'] = down_size
                self.msg['up_size'] = response['up_size']
                if down_size == response['up_size']:
                    self.msg['msg'] = 'upload3 and download3 chain is healthy'
                    self.log.log(self.msg)
                    flag = 0
                else:
                    with open('upload3.jpg', 'wb') as f:
                        f.write(data.content)
                    raise Exception('upload3 and download3 file sizes do not match')
            else:
                raise Exception('request code is not 200, download img failed')
        except Exception as e:
            self.msg['msg'] = 'download3: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                flag = self.download3(response, t)
                if not flag:
                    break
        return flag

    def upload2(self, token_info, t=2):
        response = {}
        if self.oss_url.endswith('/'):
            self.oss_url.rstrip('/')
        try:
            token = token_info['token']
            up_size = os.path.getsize(small_01_img)
            url = self.oss_url + '/upload2?size={size}&access_token={token}&expiretype={expiretype}'.format(
                **{"size": up_size, "token": token,"expiretype": expiretype})
            f = {"file": ("small_01.jpg", open(small_01_img, 'rb'), 'image/jpg', {'Expires': '0'})}
            data = requests.post(url, files=f, verify=False, timeout=10)
            status_code = data.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                data = data.json()
                response['obj_id'] = data['obj_id']
                response['token'] = token
                response['up_size'] = up_size
                self.msg['msg'] = 'upload2: upload img success'
                self.log.log(self.msg)
            else:
                raise Exception('request code is not 200')
        except Exception as e:
            self.msg['msg'] = 'upload2: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                response = self.upload2(token_info, t)
                if response:
                    break
        return response

    def download2(self, response, t=2):
        flag = 1
        if self.oss_url.endswith('/'):
            self.oss_url.rstrip('/')
        try:
            url = self.oss_url + '/files?access_token={token}&obj_id={obj_id}'.format(**response)
            data = requests.get(url, verify=False, timeout=30)
            status_code = data.status_code
            self.msg['code'] = status_code
            self.msg['obj_id'] = response['obj_id']
            if status_code == 200:
                down_size = len(data.content)
                self.msg['down_size'] = down_size
                self.msg['up_size'] = response['up_size']
                if down_size == response['up_size']:
                    self.msg['msg'] = 'upload2 and download2 chain is healthy'
                    self.log.log(self.msg)
                    flag = 0
                else:
                    with open('small_01.jpg', 'wb') as f:
                        f.write(data.content)
                    raise Exception('Upload2 and download2 file sizes do not match')
            else:
                raise Exception('request code is not 200, download img failed')
        except Exception as e:
            self.msg['msg'] = 'download2: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                flag = self.download2(response, t)
                if not flag:
                    break
        return flag

    def key_upload2(self, token_info, t=2):
        import uuid
        key = '/1/' + str(uuid.uuid4())
        response = {}
        if self.oss_url.endswith('/'):
            self.oss_url.rstrip('/')
        try:
            token = token_info['token']
            up_size = os.path.getsize(small_02_img)
            url = self.oss_url + '/upload2?size={size}&access_token={token}&expiretype={expiretype}&key={key}'.format(
                **{"size": up_size, "token": token,"expiretype": expiretype, "key": key})
            f = {"file": ("small_02.jpg", open(small_02_img, 'rb'), 'image/jpg', {'Expires': '0'})}
            data = requests.post(url, files=f, verify=False, timeout=10)
            status_code = data.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                response['key'] = key
                response['token'] = token
                response['up_size'] = up_size
                self.msg['msg'] = 'key_upload2: upload img success'
                self.log.log(self.msg)
            else:
                raise Exception('request code is not 200')
        except Exception as e:
            self.msg['msg'] = 'key_upload2: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                response = self.key_upload2(token_info, t)
                if response:
                    break
        return response

    def key_download2(self, response, t=2):
        flag = 1
        if self.oss_url.endswith('/'):
            self.oss_url.rstrip('/')
        try:
            url = self.oss_url + '/files?access_token={token}&key={key}'.format(**response)
            data = requests.get(url, verify=False, timeout=30)
            status_code = data.status_code
            self.msg['code'] = status_code
            self.msg['key'] = response['key']
            if status_code == 200:
                down_size = len(data.content)
                self.msg['down_size'] = down_size
                self.msg['up_size'] = response['up_size']
                if down_size == response['up_size']:
                    self.msg['msg'] = 'key_upload2 and key_download2 chain is healthy'
                    self.log.log(self.msg)
                    flag = 0
                else:
                    with open('small_02.jpg', 'wb') as f:
                        f.write(data.content)
                    raise Exception('key_upload2 and key_download2 file sizes do not match')
            else:
                raise Exception('request code is not 200, download img failed')
        except Exception as e:
            self.msg['msg'] = 'key_download2: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                flag = self.key_download2(response, t)
                if not flag:
                    break
        return flag

    def marwar_upload(self, token_info, sn, t=2):
        response = {}
        params = {}
        upload_time = int(time.time())
        try:
            if self.api_url.endswith('/'):
                self.api_url.rstrip('/')
            params['token'] = token_info['token']
            params['deviceID'] = sn
            params['timeStamp'] = upload_time
            params['imageID'] = params['timeStamp'] * 1000
            url = self.api_url + '/iermu/uploadImg?deviceID={deviceID}&timeStamp={timeStamp}&imageID={imageID}&access_token={token}'.format(
                **params
            )
            f = {"image": ("marwar.jpg", open(marwar_img, 'rb'), 'image/jpg', {'Expires': '0'})}

            result = requests.post(url, files=f, verify=False, timeout=30)
            status_code = result.status_code
            self.msg['code'] = status_code
            if status_code == 200:
                response['token'] = token_info['token']
                response['up_size'] = os.path.getsize(marwar_img)
                response['upload_time'] = upload_time
                self.msg['msg'] = 'upload4: upload img success'
                self.log.log(self.msg)
            else:
                self.log.log(result.text, mode=False)
                raise Exception('request code is not 200, upload img failed')
        except Exception as e:
            self.msg['msg'] = 'upload4: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                response = self.marwar_upload(token_info, sn, t)
                if response:
                    break
        return response

    def get_marwar_upload_obj_id(self, response, t=2):
        token = response['token']
        ret = {}
        try:
            if self.oss_url.endswith('/'):
                self.oss_url.rstrip('/')
            url = self.oss_url + '/fileinfo/last_objs?type=3&count=1&access_token={token}'.format(token=token)
            result = requests.get(url, timeout=30)
            status_code = result.status_code
            if status_code == 200:
                obj_infos = result.json()['obj_infos']
                for obj in obj_infos:
                    if abs(int(obj['upload_time']) - response['upload_time']) <= 30:
                        ret['obj_id'] = obj['obj_id']
                        ret['token'] = token
                        ret['up_size'] = response['up_size']
                        self.msg['msg'] = 'get_upload4_obj_id: get obj_id success'
                        self.log.log(self.msg)
                    else:
                        raise Exception('upload4 has no obj_id ,upload file failed')
            else:
                raise Exception('request code is not 200, get obj_id failed')
        except Exception as e:
            self.msg['msg'] = 'get_upload4_obj_id: ' + str(e)
            self.log.log(self.msg, mode=False)
            while t > 0:
                t -= 1
                ret = self.get_marwar_upload_obj_id(response, t)
                if ret:
                    break
        return ret

    def check_v2(self, cid):
        flag = 1
        token_info = self.get_token(cid)
        if token_info:
            response = self.upload2(token_info)
            if response:
                flag = self.download2(response)
        return flag

    def check_key_v2(self, cid):
        flag = 1
        token_info = self.get_token(cid)
        if token_info:
            response = self.key_upload2(token_info)
            if response:
                flag = self.key_download2(response)
        return flag

    def check_v3(self, cid):
        flag = 1
        token_info = self.get_token(cid)
        if token_info:
            response = self.upload3(token_info)
            if response:
                flag = self.download3(response)
        return flag

    def check_marwar(self, cid):
        flag = 1
        token_info = self.get_token(cid)
        if token_info:
            response = self.marwar_upload(token_info, config['sn'])
            if response:
                ret = self.get_marwar_upload_obj_id(response)
                if ret:
                    flag = 0
        return flag


if __name__ == '__main__':
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))
    yaml_file = os.path.join(BASE_DIR, 'config.yml')
    config = yaml.load(open(yaml_file, 'r'))['config']
    api_url = config.get('api_url', None)
    oss_url = config.get('oss_url', None)
    cid = config.get('cid', None)
    sn = config.get('sn', None)
    expiretype = config.get("expiretype")
    parser = optparse.OptionParser()
    parser.add_option('-i', '--interface', dest='interface', help='interface in ["oss","event","marwar"]')
    argv = sys.argv[1:]
    (options, args) = parser.parse_args(argv)
    cc = ChainCheck(api_url, oss_url)
    if options.interface:
        interface = options.interface
    else:
        interface = 'event'
    if interface == 'oss':
        flag = cc.check_v2(cid)
    elif interface == 'event':
        flag = cc.check_v3(cid)
    elif interface == 'marwar':
        flag = cc.check_marwar(cid)
    elif interface == "oss_key":
        flag = cc.check_key_v2(cid)
    else:
        cc.log.log('invalid arguments ', mode=False)
        flag = 1
    print flag



