#!/usr/bin/env python
# -*- coding: utf-8 -*-
import socket
from binascii import hexlify

url_prefix = ".antelopecloud.cn"
ip_list = []
with open("ip.txt","r") as f:
    for ip in f:
        if ip:
            packed_ip_addr = socket.inet_aton(ip.strip())
            hexStr = hexlify(packed_ip_addr)
            print "%s %s%s" % (ip, hexStr, url_prefix)
# for ip in ip_list:
#     packed_ip_addr = socket.inet_aton(ip)
#     hexStr = hexlify(packed_ip_addr)
#     print "%s %s%s" % (ip, hexStr, url_prefix)
