#!/usr/bin/env python
# -*- coding: utf-8 -*-
import requests

if __name__ == '__main__':
    try:
        headers = {
            'Content-Type': 'application/json'
        }
        data = requests.get('http://127.0.0.1:8230/heartbeat',headers=headers,timeout=30).json()
        print data['count']
    except Exception:
        print 49999
