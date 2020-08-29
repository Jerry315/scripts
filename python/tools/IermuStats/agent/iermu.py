# -*- coding: utf-8 -*-
import datetime
import yaml
import os
import logging
from logging.handlers import RotatingFileHandler
from rediscluster import StrictRedisCluster
from flask import Flask, jsonify, request


app = Flask(__name__)


@app.route('/v1/iermu/stats')
def iremu():
    key = "index_web.event.stats.storage.%s" % (datetime.date.today() - datetime.timedelta(days=1)).strftime('%Y%m%d')
    rd = StrictRedisCluster(startup_nodes=startup_nodes, password=password)
    limit = int(request.args.get('limit',None))
    if not limit:
        limit = 30
    data = rd.hgetall(key)
    for cid, value in data.items():
        if cid == 'date':
            continue
        if int(value) < limit:
            data.pop(cid)
    return jsonify(data)


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
        fmt = logging.Formatter(fmt="%(message)s")
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


if __name__ == '__main__':
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))
    RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'iermu_stats.access.log')
    ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'iermu_stats.error.log')
    logger = Logger()
    try:
        config_file = os.path.join(BASE_DIR, 'config.yml')
        config = yaml.load(open(config_file, 'rb'))['config']
        startup_nodes = config['redis']['startup_nodes']
        password = config['redis']['password']
        host = config['host']
        port = config['port']
        app.run(host=host,port=port)
    except Exception as e:
        logger.log(str(e),mode=False)
