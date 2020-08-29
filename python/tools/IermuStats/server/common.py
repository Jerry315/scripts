# -*- coding: utf-8 -*-
import os
import yaml
import logging
import smtplib
from logging.handlers import RotatingFileHandler
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'check_device_stats.txt')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'check_device_stats.error.log')
config_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(config_file, 'rb'))['config']


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


class SendMail(object):
    def __init__(self):
        self._smtp_server = config['smtp']['smtp_server']
        self._mail_user = config['smtp']['user']
        self._mail_passwd = config['smtp']['passwd']
        self._type = 'plain'
        self.server = smtplib.SMTP_SSL(self._smtp_server, 465)
        self.server.login(self._mail_user, self._mail_passwd)

    def send_plain(self, to, subject, msg):
        message = MIMEText(msg, self._type, 'utf-8')
        message['Subject'] = subject
        message['From'] = self._mail_user
        message['TO'] = ';'.join(to)
        self.server.sendmail(self._mail_user, to, message.as_string())

    def send_file(self,to,subject,text,files):
        message = MIMEMultipart()
        message['Subject'] = subject
        message['From'] = self._mail_user
        message['TO'] = ';'.join(to)
        message.attach(MIMEText(text,'plain','utf-8'))
        att = MIMEText(open(files,'rb').read(),'base64','utf-8')
        att["Content-Type"] = 'application/octet-stream'
        att["Content-Disposition"] = 'attachment; filename="check_device_stats.txt"'
        message.attach(att)
        self.server.sendmail(self._mail_user,to,message.as_string())