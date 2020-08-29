# -*- coding: utf-8 -*-
import os
import redis, json, optparse, sys, psutil
from common import config, Logger


class Monitor(object):
    def __init__(self, host, port, password):
        self.host = host
        self.port = port
        self.password = password
        self.rd = self.redis_handler()

    def redis_handler(self):
        rd = redis.Redis(host=self.host, port=self.port, password=self.password)
        return rd

    @property
    def memory(self):
        tmp = {}
        data = self.rd.info('Memory')
        total_system_memory = data.get("total_system_memory", None)
        if not total_system_memory:
            total_system_memory = psutil.virtual_memory().total
        tmp["{#TOTAL_MEMORY}"] = str(total_system_memory)
        tmp["{#USED}"] = str(data["used_memory"])
        print data["used_memory"]
        return tmp

    @property
    def cpu(self):
        tmp = dict()
        data = self.rd.info('CPU')
        for key, value in data.items():
            tmp["{#" + key.upper() + "}"] = str(value)
        print data['used_cpu_sys']
        return tmp

    def db(self, n):
        tmp = dict()
        data = self.rd.info('Keyspace')['db' + str(n)]
        for key, value in data.items():
            tmp["{#" + key.upper() + "}"] = str(value)
        print data['keys']
        return tmp


def query():
    data = {"data": []}
    info = os.popen("netstat -tlnp|grep redis-server | awk '{print $4}'").read().split("\n")
    for item in info:
        if item and not item.startswith(':'):
            host, port = item.split(":")
            data["data"].append({"{#REDISIP}": host, "{#REDISPORT}": port})
    print json.dumps(data)
    return data


if __name__ == '__main__':
    logger = Logger()
    try:
        password = config['redis']['password']
        usage = 'python monitor_redis.sh [-i<ip>] [-p<port>] [-cdmhq]'
        parser = optparse.OptionParser(usage)
        parser.add_option('-m', '--memory', action="store_false", help='redis memory info')
        parser.add_option('-i', '--ip', dest="ip", help='redis bind ip')
        parser.add_option('-p', '--port', dest="port", help='redis bind port')
        parser.add_option('-c', '--cpu', action="store_false", help='redis cpu info')
        parser.add_option('-d', '--db', dest="db", help='query redis db info')
        parser.add_option('-q', '--query', action="store_false", help='query redis base info')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        if '-q' in argv:
            query()
            exit()
        db = options.db
        if db:
            db = int(db)
        else:
            db = config['redis']['db']
        ip = options.ip
        port = options.port
        if port:
            port = int(port)
        if ip and port:
            monitor = Monitor(ip, port, password)
            if '-m' in argv:
                data = monitor.memory
            elif '-c' in argv:
                data = monitor.cpu
            elif '-d' in argv:
                data = monitor.db(db)
        else:
            parser.print_help()
    except Exception as e:
        logger.error(str(e))
