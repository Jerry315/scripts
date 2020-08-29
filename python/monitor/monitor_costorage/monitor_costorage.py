# -*- coding: utf-8 -*-
import time
import json
import socket
import optparse
import sys
import os
import yaml
import subprocess
from logger import Logger, BASE_DIR

try:
    import xml.etree.cElementTree as ET
except ImportError:
    import xml.etree.ElementTree as ET


def xml_parse(xfile):
    data = []
    tree = ET.parse(xfile)
    root = tree.getroot()
    for partition in root[0].findall("partition"):
        tmp = dict()
        tmp['enable'] = partition.find('enable').text
        tmp['public_ip'] = partition.find('public_ip').text
        tmp['local_ip'] = partition.find('local_ip').text
        tmp['upload_port'] = partition.find('upload_port').text
        tmp['query_port'] = partition.find('query_port').text
        data.append(tmp)
    return data


def query(xfile):
    data = xml_parse(xfile)
    result = dict()
    en0 = 0
    en0_ports = []
    en1 = 0
    en1_ports = []
    upload_ports = []
    for item in data:
        if int(item['enable']) == 1:
            upload_ports.append({"{#UPLOAD_PORT}": item['upload_port']})
            en1_ports.append(item['upload_port'])
            en1 += 1
        else:
            en0 += 1
            en0_ports.append(item['upload_port'])
    logger.log(json.dumps({
        "timestamp": int(time.time()),
        "log_type": "costorage",
        "hostname": socket.gethostname(),
        "enable_1_count": en1,
        "enable_1_ports": en1_ports,
        "enable_0_count": en0,
        "enable_0_ports": en0_ports
    }))

    result['data'] = upload_ports

    print json.dumps(result)


def online(xfile):
    data = xml_parse(xfile)
    count = 0
    # result = dict()
    cos = subprocess.Popen('ps -ef |grep "./COStorage" | grep -v grep | wc -l', stdin=subprocess.PIPE,
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
    proc = cos.stdout.read()
    for item in data:
        if int(item['enable']) == 1:
            count += 1
    if count:
        count += 1
    # result["data"] = [{"{#ONLINE}": str(count)}, {"{PROC}": proc.strip('\n')}]
    if int(proc) != count:
        print 0
    else:
        print 1
    # print json.dumps(result)


def check(port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = sock.connect_ex(('127.0.0.1', port))
    if result == 0:
        print 1
    else:
        print 0


if __name__ == '__main__':
    logger = Logger()
    try:
        yaml_file = os.path.join(BASE_DIR, 'config.yml')
        config = yaml.load(open(yaml_file, 'rb'))['config']
        usage = 'python monitor_costrage.py [-q] [-p<port>]'
        parser = optparse.OptionParser(usage)
        parser.add_option('-p', '--port', dest='port', help='costorage upload port')
        parser.add_option('-q', '--query', action="store_false", help='query costorage info')
        parser.add_option('-n', '--online', action="store_false", help='query costorage process')
        argv = sys.argv[1:]
        if '-q' in argv:
            query(config['costorage']['config'])
            exit()
        if '-n' in argv:
            online(config['costorage']['config'])
        (options, args) = parser.parse_args(argv)
        port = int(options.port)
        check(port)
    except Exception as e:
        logger.log(str(e), mode=False)
