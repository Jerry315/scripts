# -*- coding: utf-8 -*-
import os
import logging
import requests
import json
import yaml
from logging.handlers import RotatingFileHandler
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'monitor_diskgroup.access.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'monitor_diskgroup.error.log')
AGENT_RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'diskgroup_rate.access.log')
AGENT_ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'diskgroup_rate.error.log')
yaml_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(yaml_file, 'rb'))['config']


class BaseLogger(object):
    __instance = None
    _run_log_file = None
    _error_log_file = None

    def __init__(self):
        self.run_log_file = self._run_log_file
        self.error_log_file = self._error_log_file
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
            self.error_logger.error(message, exc_info=True)


def get_data(url):
    '''
    :param url: 磁盘调度HTTP接口
    :return:
    '''
    response = {
        'code': None,
        'msg': None
    }
    data = None
    result = requests.get(url)
    if result.status_code == 200:
        data = result.json()
        response['code'] = 200
        response['msg'] = '[get_data]: get disk data success.'
        logger.log(response)
    else:
        response['code'] = result.status_code
        response['msg'] = '[get_data]: get disk data failed'
        logger.log(response, mode=False)
    return data


def get_group_id(url,func):
    data = func(url)
    group_id_list = []
    group_data = data.get('monitor_info', None)
    if group_data:
        for item in group_data:
            if item['group_id'] == 0:
                continue
            group_id_list.append({
                "{#GROUPID}": str(item['group_id']),
                "{#MAX_UPLOAD_RATE}": str(item['max_upload_rate']),
                "{#LIMIT_CAPACITY}": str((int(item['size'] / 1024 / 1024 / 1024 / 1024)/2)+1)+'T'
            })
    return json.dumps({"data": group_id_list})


class Logger(BaseLogger):
    _run_log_file = RUN_LOG_FILE
    _error_log_file = ERROR_LOG_FILE


class AgentLogger(BaseLogger):
    _run_log_file = AGENT_RUN_LOG_FILE
    _error_log_file = AGENT_ERROR_LOG_FILE

    def initialize_run_log(self):
        self.check_path_exist(self.run_log_file)
        file_1_1 = RotatingFileHandler(filename=self.run_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(message)s")
        file_1_1.setFormatter(fmt=fmt)
        logger1 = logging.Logger('run_log', level=logging.INFO)
        logger1.addHandler(file_1_1)
        self.run_logger = logger1

logger = Logger()