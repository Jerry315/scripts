# -*- coding: utf-8 -*-
import os
import logging
from logging.handlers import RotatingFileHandler

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'monitor_costorage.access.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'monitor_costorage.error.log')


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

class Logger(BaseLogger):
    _run_log_file = RUN_LOG_FILE
    _error_log_file = ERROR_LOG_FILE