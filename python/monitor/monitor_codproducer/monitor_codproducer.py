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
        result["data"].append({"{#CONF}": "", "{#RPCPORT}": "1", "{#QUERYPORT}": "1"})
    else:
        for f in files:
            general = yaml.load(open(f, 'rb'))['general']
            tmp = {"{#CONF}": f}
            if general.get("rpcPort", None):
                tmp["{#RPCPORT}"] = general["rpcPort"].split(":")[1]
            else:
                tmp["{#RPCPORT}"] = ""
            if general.get("queryport", None):
                tmp["{#QUERYPORT}"] = general["queryport"].strip(":")
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
            host = os.popen("netstat -anp|grep %s | grep codproducer | awk '{print $4}'" % port).read().split(':')[0]
            if not host:
                host = '127.0.0.1'
            result = sock.connect_ex((host, port))
            if result == 0:
                pid = os.popen("netstat -anp|grep %s | grep codproducer | awk '{print $7}'" % port).read().split('/')[0]
                if not pid:
                    flag = 0
            else:
                flag = 0
    print flag


if __name__ == '__main__':
    logger = Logger()
    try:
        conf_dir = config["codproducer"]["conf_dir"]
        usage = 'python monitor_codispatcher.py [-q] [-p<port>]'
        parser = optparse.OptionParser(usage)
        parser.add_option('-p', '--port', dest='port', help='codproducer services port')
        parser.add_option('-q', '--query', action="store_false", help='query codproducer service is up')
        parser.add_option('-c', '--check', action="store_false", help='query codproducer config file is exist')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        if '-q' in argv:
            parse_yaml(conf_dir)
            exit()
        ports = options.port.strip(',').split(',')
        check(ports)
    except Exception as e:
        logger.error(str(e))
