# -*- coding: utf-8 -*-
import logging,os
import smtplib
import yaml
from logging.handlers import RotatingFileHandler
from email.mime.text import MIMEText

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'check_cid_timout.access.log')
UNNORMAL_CID_FILE = os.path.join(BASE_DIR, 'log', 'unnormal_cid_file.txt')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'check_cid_timout.error.log')
config_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(config_file, 'rb'))['config']


class Logger(object):
    __instance = None

    def __init__(self):
        self.run_log_file = RUN_LOG_FILE
        self.error_log_file = ERROR_LOG_FILE
        self.unnormal_cid_file = UNNORMAL_CID_FILE
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
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('run_log', level=logging.INFO)
        logger1.addHandler(file_1_1)
        self.run_logger = logger1

    def initialize_normal_logger(self, level):
        self.check_path_exist(self.run_log_file)
        file_1_1 = RotatingFileHandler(filename=self.run_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('run_log', level=level)
        logger1.addHandler(file_1_1)
        return logger1

    def initialize_unnormal_log(self, level):
        self.check_path_exist(self.unnormal_cid_file)
        file_1_1 = RotatingFileHandler(filename=self.unnormal_cid_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s  - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('error_log', level=level)
        logger1.addHandler(file_1_1)
        return logger1

    def initialize_error_log(self):
        self.check_path_exist(self.error_log_file)
        file_1_1 = RotatingFileHandler(filename=self.error_log_file, maxBytes=1024 * 1024 * 2, backupCount=15,
                                       encoding='utf-8')
        fmt = logging.Formatter(fmt="%(asctime)s  - %(levelname)s :  %(message)s")
        file_1_1.setFormatter(fmt)
        logger1 = logging.Logger('error_log', level=logging.ERROR)
        logger1.addHandler(file_1_1)
        self.error_logger = logger1

    def debug(self, msg):
        logger = self.initialize_normal_logger(logging.DEBUG)
        logger.debug(msg)

    def info(self, msg):
        logger = self.initialize_normal_logger(logging.INFO)
        logger.info(msg)

    def warn(self, msg):
        logger = self.initialize_unnormal_log(logging.WARNING)
        logger.warning(msg)

    def error(self, msg):
        self.error_logger.error(msg, exc_info=True)

    def log(self, message, mode=True):
        """
        写入日志
        :param message: 日志信息
        :param mode: True表示运行信息，False表示错误信息
        :return:
        """
        if mode:
            self.run_logger.info(message)
        else:
            self.error_logger.error(message, exc_info=True)


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
