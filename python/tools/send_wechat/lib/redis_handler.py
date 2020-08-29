#!/usr/bin/env python
# -*- coding: utf-8 -*-
from redis import Redis

from settings import redis_host, redis_passwd, redis_port, redis_db

rd = Redis(host=redis_host, password=redis_passwd, port=redis_port, db=redis_db)

# rd.setex("token", "xxxxxx", 7200)
# rd.setex("timer", 60, 60)
# rd.set("counter", 1)
# timer = rd.get("timer")
# print timer
