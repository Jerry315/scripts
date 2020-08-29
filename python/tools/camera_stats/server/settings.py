#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
from threadpool import ThreadPool


BASE_DIR = os.path.dirname(os.path.abspath(__file__))
RUN_LOG_FILE = os.path.join(BASE_DIR, 'log', 'camera_stats.access.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'log', 'camera_stats.error.log')
config_file = os.path.join(BASE_DIR, 'config.yml')
config = yaml.load(open(config_file, 'rb'))["config"]
cid_files = os.path.join(BASE_DIR, 'log', 'cids')
report_file = os.path.join(BASE_DIR, "log", "camera_stats_report.xls")
zip_name = os.path.join(BASE_DIR, "log", "camera_stats_report.zip")
pool = ThreadPool(config['pool_size'])
