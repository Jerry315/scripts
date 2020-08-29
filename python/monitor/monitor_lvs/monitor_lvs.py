# -*- coding: utf-8 -*-
import os
import json
import re
import optparse
import sys
import time
from common import Logger


# def query():
#     """
#     获取当前机器上lvs的基础信息，vip和port
#     :return:
#     """
#     info = os.popen("ipvsadm -Ln").read().split("\n")
#     data = {"data": []}
#     for item in info[3:]:
#         if item.startswith("TCP"):
#             ip, port = re.split(r"\s+",item)[1].split(':')
#             data["data"].append({"{#IP}": ip, "{#PORT}": port})
#
#     print json.dumps(data)

def parse_data():
    """
    解析ipvsadm -Ln命令的数据，得到两个字典，一个是vip对应后端的真实ip和端口，一个是后端真实ip端口以及连接数
    :return:
    """
    data = {"create_time": int(time.time()),"data": []}
    info = os.popen("ipvsadm -Ln").read().split("\n")
    tmp = dict()
    for item in info[3:]:
        if not item:
            continue
        if item.startswith("TCP"):
            if tmp:
                data["data"].append(tmp)
            tmp = dict()
            vip,port = re.split(r"\s+", item)[1].split(":")
            tmp["vip"] = vip
            tmp['port'] = port
            tmp['backend'] = []
            continue
        else:
            _, hosts, _, _, ActiveConn, InActConn = re.split(r"\s+", item.strip())
            host,port = hosts.split(":")
            tmp['backend'].append({"host": host,"port": port,"ActiveConn":ActiveConn,"InActConn": InActConn})
    if tmp:
        data["data"].append(tmp)
    logger.info(json.dumps(data))


if __name__ == '__main__':
    logger = Logger()
    try:
        parse_data()
    except Exception as e:
        logger.error(str(e))
