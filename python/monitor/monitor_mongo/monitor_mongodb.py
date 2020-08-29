# -*- coding: utf-8 -*-
import click
import json
import os
from pymongo import MongoClient
from common import config, Logger


def convert_dict_to_str(item, prefix=None):
    """把字典转换成字符串"""
    ret = ''
    for k, v in item.items():
        if ret:
            ret += ','
        if isinstance(v, dict):
            if prefix:
                t = convert_dict_to_str(v, prefix=prefix + '.' + k)
            else:
                t = convert_dict_to_str(v, prefix=k)
            ret += t
        else:
            if prefix:
                ret += '{}: {}'.format(prefix + '.' + k, v)
            else:
                ret += '{}: {}'.format(k, v)
    return ret


def mongo_conn(port):
    port = int(port)
    mongo_url = mongo_conf[port]["url"]
    conn = MongoClient(mongo_url, connect=False)
    return conn


def gatherServerStatusMetrics(port):
    conn = mongo_conn(port)
    serverMetrics = dict()
    serverStatus = conn.admin.command("serverStatus")
    if "asserts" in serverStatus:
        serverMetrics['asserts'] = serverStatus['asserts']
    if "connections" in serverStatus:
        serverMetrics['connections'] = serverStatus['connections']
    if 'mem' in serverStatus:
        serverMetrics["memory"] = serverStatus["mem"]
    if "network" in serverStatus:
        serverMetrics['network'] = serverStatus['network']
    if "repl" in serverStatus:
        serverMetrics['repl'] = serverStatus['repl']
    if "extra_info" in serverStatus:
        serverMetrics['extra_info'] = serverStatus['extra_info']
    if "activeClients" in serverStatus["globalLock"]:
        serverMetrics["active_clients"] = serverStatus["globalLock"]["activeClients"]
    if 'opcounters' in serverStatus:
        serverMetrics["opcounters"] = serverStatus['opcounters']
    if 'opcountersRepl' in serverStatus:
        serverMetrics["opcountersRepl"] = serverStatus['opcountersRepl']
    if "uptime" in serverStatus:
        serverMetrics["uptime"] = serverStatus["uptime"]
    if "version" in serverStatus:
        serverMetrics["version"] = serverStatus["version"]
    if "process" in serverStatus:
        serverMetrics["process"] = serverStatus["process"]
    if "globalLock" in serverStatus:
        serverMetrics["globalLock"] = serverStatus["globalLock"]
    return serverMetrics


def gatherReplSetGetstatus(port):
    conn = mongo_conn(port)
    return conn.admin.command("replSetGetStatus")


def parse_metric(serverMetrics, metric):
    tmp = dict()
    data = serverMetrics[metric]
    for key, value in data.items():
        if isinstance(value, float):
            value = '%.2f' % value
        elif not isinstance(value, bool):
            value = str(value)
        tmp[key] = value
    return tmp


@click.group()
def cli():
    pass


@click.command()
@click.argument("port")
def replhealth(port):
    '''
    获取集群状态
    port: 数据库监听端口
    '''
    flag = 1
    conn = mongo_conn(port)
    data = conn.admin.command("replSetGetStatus")
    for i in range(len(data['members'])):
        if data['members'][i]['health'] == 1:
            continue
        else:
            flag = 0
    print flag


@click.command()
@click.argument("config")
def getpid(config):
    '''
    根据配置文件获取进程状态
    config: mongodb实例的配置文件全路径
    '''
    pid = \
        os.popen("ps -ef | grep %s | grep -v grep | grep mongod | awk '{print $2}'" % config).read().strip(
            "\n").split(
            "\n")[1]
    if pid:
        print 1
    else:
        print 0


@click.command()
@click.argument("port")
def replrelay(port):
    '''
    获取复制延时时间
    port: 数据库监听端口
    '''
    primary_optime = 0
    secondary_optime = 0
    conn = mongo_conn(port)
    data = conn.admin.command("replSetGetStatus")
    for key in data["members"]:
        if key['stateStr'] == 'SECONDARY':
            secondary_optime = key['optimeDate']
        if key['stateStr'] == 'PRIMARY':
            primary_optime = key['optimeDate']
    seconds_lag = (primary_optime - secondary_optime).total_seconds()
    print str(seconds_lag)


@click.command()
@click.argument("port")
def op(port):
    '''
    获取当前在执行任务的数目
    port: 数据库监听端口
    '''
    conn = mongo_conn(port)
    data = conn.admin.current_op()
    print len(data['inprog'])
    for item in data['inprog']:
        if item['secs_running'] >= config['timeout']:
            ret = convert_dict_to_str(item)
            logger.warn(ret)


@click.command()
def query():
    '''
    查询本机运行的mongodb信息
    '''
    data = {"data": []}
    for port in mongo_conf:
        pid = os.popen("netstat -tlnp|grep %s | grep  mongod| awk '{print $7}'" % port).read().split('/')[0]
        if pid:
            conf_file = os.popen("ps -ef | grep %s | grep -v grep | awk '{print $10}'" % pid).read()
            data["data"].append({"{#PORT}": str(port), "{#CONF_FILE}": conf_file.strip().strip('\n')})
    print json.dumps(data)


@click.command()
@click.argument("port")
def uptime(port):
    '''
    获取数据库启动运行时间（s）
    port: 数据库监听端口
    '''
    ServerMetrics = gatherServerStatusMetrics(port)
    print int(ServerMetrics["uptime"])


@click.command()
@click.argument("port")
def version(port):
    '''
    获取mongodb版本信息
    port: 数据库监听端
    '''
    ServerMetrics = gatherServerStatusMetrics(port)
    print ServerMetrics["version"]


@click.command()
@click.argument("port")
def process(port):
    """
    获取进程名称
    port: 数据库监听端
    """
    ServerMetrics = gatherServerStatusMetrics(port)
    print ServerMetrics["process"]


@click.command()
@click.argument("port")
@click.argument("item")
@click.argument("metric")
def globallock(port, item, metric):
    """
    port: 数据库监听端口
    item: currentQueue全局锁总数、读写数；activeClients活跃客户端数量
    metric: total、readers、writers
    """
    ServerMetrics = gatherServerStatusMetrics(port)
    print ServerMetrics["globalLock"][item][metric]


@click.command()
@click.argument("port")
@click.argument("item")
@click.argument('metric', default=None)
def common(port, item, metric):
    '''
    公共接口，可以获取大部分信息
    port: 数据库监听端口
    item: 要查询的条目
    metric: 要查询的子条目
    '''
    ServerMetrics = gatherServerStatusMetrics(port)
    data = parse_metric(ServerMetrics, item)
    if metric:
        print data[metric]
    else:
        print data


if __name__ == '__main__':
    logger = Logger()
    try:
        mongo_conf = config["mongodb"]["instance"]
        cli.add_command(common)
        cli.add_command(query)
        cli.add_command(op)
        cli.add_command(replrelay)
        cli.add_command(replhealth)
        cli.add_command(uptime)
        cli.add_command(version)
        cli.add_command(process)
        cli.add_command(globallock)
        cli.add_command(getpid)
        cli()
    except Exception as e:
        logger.error(str(e))
