# -*- coding: utf-8 -*-
import os
import yaml
import logging
import smtplib
from logging.handlers import RotatingFileHandler
from email.mime.text import MIMEText
from schematics.models import Model

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'camera_stats.access.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'camera_stats.error.log')
config_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(config_file, 'rb'))['config']


class Logger(object):
    __instance = None

    def __init__(self):
        self.run_log_file = RUN_LOG_FILE
        self.error_log_file = ERROR_LOG_FILE
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

    def debug(self, msg):
        logger = self.initialize_run_log(logging.DEBUG)
        logger.debug(msg)

    def info(self, msg):
        logger = self.initialize_run_log(logging.INFO)
        logger.info(msg)

    def warn(self, msg):
        logger = self.initialize_run_log(logging.WARNING)
        logger.warning(msg)

    def error(self, msg):
        logger = self.initialize_error_log(logging.ERROR)
        logger.error(msg, exc_info=True)


class SendMail(object):
    def __init__(self):
        self._smtp_server = config['smtp']['smtp_server']
        self._mail_user = config['smtp']['user']
        self._mail_passwd = config['smtp']['passwd']
        self._type = 'plain'
        self.server = smtplib.SMTP_SSL(self._smtp_server, 465)
        self.server.login(self._mail_user, self._mail_passwd)

    def send(self, to, subject, msg):
        message = MIMEText(msg, self._type, 'utf-8')
        message['Subject'] = subject
        message['From'] = self._mail_user
        message['TO'] = ';'.join(to)
        self.server.sendmail(self._mail_user, to, message.as_string())


class BaseModel(Model):
    __database__ = ''
    __collection__ = ''

    def __init__(self, *args, **kwargs):
        super(BaseModel, self).__init__(*args, strict=False, **kwargs)
        self._collection = self.__database__.get_collection(self.__collection__)

    @classmethod
    def get_collection(cls):
        '''获取collection对象'''
        return cls.__database__.get_collection(cls.__collection__)

    @classmethod
    def find(cls, filter, fields=None, **kwargs):
        if fields:
            fields = dict([(f, True) for f in fields])
        collection = cls.get_collection()
        records = collection.find(filter, fields, **kwargs)
        return records

