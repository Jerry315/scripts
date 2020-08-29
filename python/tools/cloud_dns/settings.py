#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
from aliyunsdkcore.client import AcsClient


BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


###############aliyun arm info###############
AK = "xxxx"
SECRET = "xxxx"
REGION_ID = "default"
ENDPOINT = "alidns.aliyuncs.com"
FORMAT = "JSON"
VERSION = "2015-01-09"
PROTOCOL = "https"
METHOD = "POST"

###############aliyun request client #######
CLIENT = AcsClient(AK, SECRET, REGION_ID)

##############log file#######################
RUN_LOG_FILE = os.path.join(BASE_DIR, 'logs', 'cloud_dns.access.log')
ERROR_LOG_FILE = os.path.join(BASE_DIR, 'logs', 'cloud_dns.error.log')

##############records file###################
RECORDS_FILE = os.path.join(BASE_DIR,'records.xlsx')
