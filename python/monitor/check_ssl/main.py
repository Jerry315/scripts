# -*- coding: utf-8 -*-
import socket
import sys
import ssl
import time
import datetime
import json
import optparse
from prettytable import PrettyTable
from prettytable import ALL
from common import config, Logger


def check_ssl(hostname, host, port, expire=60):
    msg = {
        "hostname": hostname,
        "host": host,
        "port": port,
        "expire": expire,
        "expire_datetime": None,
        "status": False
    }
    ctx = ssl.create_default_context()
    s = ctx.wrap_socket(socket.socket(), server_hostname=hostname) #
    s.connect((host, port))
    cert = s.getpeercert()
    today = time.time()
    expire_time = 0
    expire_date = datetime.datetime.today()

    for key, value in cert.items():
        if key == 'notAfter':
            expire_date = value.rstrip("GMT").strip()
            expire_date = datetime.datetime.strptime(expire_date, '%b %d %H:%M:%S %Y') + datetime.timedelta(hours=8)
            expire_time = time.mktime(time.strptime(expire_date.strftime('%Y-%m-%d %H:%M:%S'), '%Y-%m-%d %H:%M:%S'))
    msg['expire_datetime'] = expire_date.strftime('%Y-%m-%d %H:%M:%S')
    msg['expire_day'] = (expire_date - datetime.datetime.today()).days
    if (expire_time - today) / 86400 >= expire:
        msg["status"] = True
    logger.info(json.dumps(msg))
    return msg


def query():
    try:
        tmp = {}
        for project in config['project']:
            urls = config['project'][project]['urls']
            for item in urls:
                url = item['url']
                for host in item['hosts']:
                    msg = check_ssl(url, host['ip'], host['port'], expire=expire)
                    if url in tmp:
                        tmp[url].append(msg)
                    else:
                        tmp[url] = []
                        tmp[url].append(msg)
        tb = PrettyTable(["HostName", "Host", "Expire", "ExpireDateTime", "ExpireDay"])
        tb.align = 'c'
        tb.valign = 'm'
        tb.hrules = ALL
        for url, value in tmp.items():
            tmp_hosts = []
            tmp_expire = []
            tmp_expire_date_time = []
            tmp_expire_day = []
            for msg in value:
                tmp_hosts.append(msg['host'])
                tmp_expire.append(msg['expire'])
                tmp_expire_date_time.append(msg['expire_datetime'])
                tmp_expire_day.append(msg['expire_day'])
            tb.add_row([
                url,
                '\n'.join(tmp_hosts),
                '\n'.join([str(n) for n in tmp_expire]),
                '\n'.join(tmp_expire_date_time),
                '\n'.join([str(n) for n in tmp_expire_day])
            ])
        print(tb.get_string(align="l"))
    except Exception as e:
        logger.error(str(e))


def parse_url():
    result = dict()
    result['data'] = []
    for project in config['project']:
        for item in config['project'][project]['urls']:
            for host in item["hosts"]:
                result['data'].append(
                    {"{#URL}": item["url"],"{#IP}": host["ip"], "{#PORT}": str(host['port'])}
                )
    print(json.dumps(result))


def zabbix(hostname, host, port, expire=60):
    try:
        msg = check_ssl(hostname, host, port, expire=expire)
        print(msg['expire_day'])
    except Exception as e:
        logger.error(str(e))
        print(0)


if __name__ == '__main__':
    logger = Logger()
    usage = "python stats.py -u www.baidu.com -j [-i<ip>] [-p<port>] [-q|-z]"
    parser = optparse.OptionParser(usage)
    parser.add_option('-e', '--expire', dest='expire', help='How many days expire ')
    parser.add_option('-i', '--ip', dest='ip', help='ipv4 address')
    parser.add_option('-p', '--port', dest='port', help='https port,default 443')
    parser.add_option('-q', '--query', dest='query', action="store_false", help="parse url to real host")
    parser.add_option('-u', '--url', dest='url', help='http url')
    parser.add_option('-z', '--zabbix', dest='zabbix', action="store_false", help='just support for zabbix')
    argv = sys.argv[1:]
    (options, args) = parser.parse_args(argv)
    ip = options.ip
    port = options.port
    url = options.url
    expire = options.expire
    try:
        if expire:
            expire = int(expire)
        else:
            expire = config['expire']
        if port:
            port = int(port)
        else:
            port = 443
        if not len(argv):
            query()
            exit()
        if ip and port and url:
            if '-q' in argv:
                print(check_ssl(url, ip, port, expire))
            elif '-z' in argv:
                zabbix(url, ip, port, expire=expire)
        else:
            if '-q' in argv:
                parse_url()
    except Exception as e:
        logger.error(str(e))
