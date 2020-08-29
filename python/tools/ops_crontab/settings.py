# -*- coding: utf-8 -*-
import os
import yaml

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
config = yaml.load(open(os.path.join(BASE_DIR, 'config.yml'), 'rb'))['ops_crontab']
if not config["log_dir"]:
    log_dir = os.path.join(BASE_DIR, "logs")
else:
    log_dir = os.path.join(BASE_DIR, config['log_dir'])
