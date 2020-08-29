#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
import random
import time
import datetime
from pytz import UTC

from threading import Thread, Semaphore

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
config_file = os.path.join(BASE_DIR, 'mutil_log.yaml')
config = yaml.load(open(config_file, 'rb'))
LOG_FILE = os.path.join(BASE_DIR, config["log"]["name"])

log_level = config["log_level"]
log_types = config["log_type"]
ip = config["ip"]
_str = config["str"]
hostname = "mutil-log"

msg = "[%s] [%s] [%s] - [ip: %s, log_type: %s, message: %s, pid: %d]\n"

EIGHT = datetime.timedelta(hours=8)

class CustomUTC(datetime.tzinfo):
    zone = "UTC"
    _utcoffset = datetime.timedelta(0)
    _dst = EIGHT
    _tzname = datetime.timedelta(0)

    def fromutc(self, dt):
        if dt.tzinfo is None:
            return self.localize(dt)
        return super(utc.__class__, self).fromutc(dt)

    def utcoffset(self, dt):
        return EIGHT

    def tzname(self, dt):
        return "UTC"

    def dst(self, dt):
        return EIGHT

    def localize(self, dt, is_dst=False):
        '''Convert naive time to local time'''
        if dt.tzinfo is not None:
            raise ValueError('Not naive datetime (tzinfo is already set)')
        return dt.replace(tzinfo=self)

    def normalize(self, dt, is_dst=False):
        '''Correct the timezone information on the given datetime'''
        if dt.tzinfo is self:
            return dt
        if dt.tzinfo is None:
            raise ValueError('Naive time - no tzinfo set')
        return dt.astimezone(self)

    def __repr__(self):
        return "<UTC>"

    def __str__(self):
        return "UTC"


utc = CustomUTC()


def date_str():
    return datetime.datetime.fromtimestamp(time.time(), tz=utc).isoformat(sep="T")


def create_log():
    sm.acquire()
    try:
        f = open(LOG_FILE, "a")
    except Exception:
        f = open(LOG_FILE, "w")
    log_msg = msg % (
        date_str(),
        hostname,
        log_level[random.randrange(len(log_level))],
        ip[random.randrange(len(ip))],
        log_types[random.randrange(len(log_types))],
        _str[random.randrange(len(_str))],
        random.randint(0, 65535))
    f.write(log_msg)
    f.close()
    sm.release()


if __name__ == '__main__':
    sm = Semaphore(config["thread"])
    while True:
        time.sleep(0.3)
        t = Thread(target=create_log)
        t.start()
