# -*- coding: utf-8 -*-
import json
import socket
import optparse
import sys
import os
import yaml
from common import config, Logger


def get_conf_file(conf_dir):
    '''
    获取配置文件
    :param conf_dir:
    :return:
    '''
    tmp = []
    try:
        if os.path.exists(conf_dir):
            f = os.listdir(conf_dir)
            for i in f:
                if i.endswith(".yaml"):
                    tmp.append(os.path.join(conf_dir, i))
    except Exception as e:
        logger.error(str(e))
    return tmp


def parse_yaml(conf_dir):
    files = get_conf_file(conf_dir)
    result = {"data": []}
    if not files:
        result["data"].append({"{#CONF}": "", "{#DEBUGPORT}": "1", "{#QUERYPORT}": "1"})
    else:
        for f in files:
            web = yaml.load(open(f, 'rb'))['web']
            tmp = {"{#CONF}": f}
            if web.get("debugPort", None):
                tmp["{#DEBUGPORT}"] = web["debugPort"].strip(":")
            else:
                tmp["{#DEBUGPORT}"] = ""
            if web.get("queryport", None):
                tmp["{#QUERYPORT}"] = web["queryport"].strip(":")
            else:
                tmp["{#QUERYPORT}"] = ""
            result["data"].append(tmp)
    print json.dumps(result)


def check(ports):
    flag = 1
    for port in ports:
        port = int(port)
        if port == 1:
            flag = -1
        else:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            host = os.popen("netstat -anp|grep %s | grep codispatcher | awk '{print $4}'" % port).read().split(':')[0]
            if not host:
                host = '127.0.0.1'
            result = sock.connect_ex((host, port))
            if result == 0:
                pid = os.popen("netstat -anp|grep %s | grep codispatcher | awk '{print $7}'" % port).read().split('/')[
                    0]
                if not pid:
                    flag = 0
            else:
                flag = 0
    print flag


if __name__ == '__main__':
    logger = Logger()
    try:
        conf_dir = config["codispatcher"]["conf_dir"]
        usage = 'python monitor_codispatcher.py [-q] [-p<port>]'
        parser = optparse.OptionParser(usage)
        parser.add_option('-p', '--port', dest='port', help='codispatcher services port')
        parser.add_option('-q', '--query', action="store_false", help='query codispatcher service is up')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        if '-q' in argv:
            parse_yaml(conf_dir)
            exit()
        ports = options.port.strip(',').split(',')
        check(ports)
    except Exception as e:
        logger.error(str(e))
