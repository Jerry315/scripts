#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import yaml
import click
from datetime import timedelta, datetime
import requests
import json

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
config_file = os.path.join(BASE_DIR, 'config.yaml')
config = yaml.load(open(config_file, 'rb'))

es_conf = config['elastic']
_formats = {
    "ymd": "%Y%m%d",
    "y.m.d": "%Y.%m.%d",
    "y-m-d": "%Y-%m-%d"
}
headers = {
    'Content-Type': "application/json",
}
log_dir = os.path.join(BASE_DIR, 'logs')
snapshot_log_file = os.path.join(log_dir, 'snapshot.%s.log' % datetime.today().strftime("%Y%m%d"))
delete_log_file = os.path.join(log_dir, 'delete.%s.log' % datetime.today().strftime("%Y%m%d"))
tag_log_file = os.path.join(log_dir, 'tag.%s.log' % datetime.today().strftime("%Y%m%d"))


class EsHandler(object):
    """docstring for SnapShot"""

    def __init__(self, url, timeout):
        self.url = url
        self.timeout = timeout

    def put_repository(self, date_fmt, max_snapshot_bytes_per_sec, max_restore_bytes_per_sec):
        url = "%s/_snapshot/repository_%s?pretty=true" % (
            self.url, date_fmt)
        data = {
            "type": "fs",
            "settings": {
                "location": "/opt/es_backup/" + date_fmt,
                "max_snapshot_bytes_per_sec": max_snapshot_bytes_per_sec,
                "max_restore_bytes_per_sec": max_restore_bytes_per_sec
            }
        }
        response = requests.put(url, data=json.dumps(data), headers=headers, timeout=self.timeout)
        write_log(snapshot_log_file, "PUT repository", response, data)
        return response

    def get_repository(self):
        response = requests.get(self.url, timeout=self.timeout)
        return response

    def delete_repository(self, indices):
        url = "%s/%s" % (self.url, indices)
        response = requests.delete(url, timeout=self.timeout)
        write_log(delete_log_file, "DELETE repository", response, data=None)
        return response

    def put_snapshot(self, index, date_fmt):
        url = "%s/_snapshot/repository_%s/snapshot?pretty=true" % (
            self.url, date_fmt)
        data = {
            "indices": index,
            "ignore_unavailable": True,
            "include_global_state": False
        }
        response = requests.put(url, data=json.dumps(data), headers=headers, timeout=self.timeout)
        write_log(snapshot_log_file, "PUT snapshot", response, data)

    def get_snapshot(self, date_fmt):
        url = "%s/_snapshot/repository_%s/snapshot" % (self.url, date_fmt)
        response = requests.get(url, timeout=self.timeout)
        write_log(snapshot_log_file, "GET snapshot", response, data=None)
        return response

    def delete_snapshot(self, date_fmt):
        url = "%s/_snapshot/repository_%s/snapshot" % (self.url, date_fmt)
        response = requests.delete(url, timeout=self.timeout)
        write_log(delete_log_file, "DELETE snapshot", response, data=None)
        return response

    def put_tag(self, indices, tag):
        for indice in indices:
            url = "%s/%s/_settings" % (self.url, indice)
            data = {"index.routing.allocation.require.tag": tag}
            response = requests.put(url, data=json.dumps(data), headers=headers, timeout=self.timeout)
            write_log(tag_log_file, "PUT tag", response, data)


def convert_date(td, fmt):
    '''根据delay_day转换成指定日期格式'''
    return (datetime.now() - timedelta(days=td)).strftime(fmt)


def es_handler():
    timeout = es_conf['timeout']
    es_url = es_conf['url']
    es_handle = EsHandler(es_url, timeout)
    return es_handle


def parse_index(indexs):
    '''生成delay_day和索引关系字典便于后续处理'''
    tmp = dict()
    for index in indexs:
        if index['enable'] != 0:
            if index['delay_days'] not in tmp:
                tmp[index['delay_days']] = []
            fmt = _formats.get(index['date_fmt'], "%Y.%m.%d")
            index_date_fmt = convert_date(index['delay_days'], fmt)
            index_list = ['%s%s' % (indice, index_date_fmt) for indice in index['index']]
            tmp[index['delay_days']] += index_list
    return tmp


def write_log(log_file, method, response, data):
    '''记录操作日志以及操作返回结果'''
    if not os.path.exists(log_dir):
        os.makedirs(log_dir)
    msg = "-*-" * 20 + " " * 4 + method + " Start " + " " * 4 + "-*-" * 20 + "\n"
    msg = msg + "url: " + str(response.url) + "\n"
    msg = msg + "data:" + str(json.dumps(data, indent=4)) + "\n"
    msg = msg + "status_code: " + str(response.status_code) + "\n"
    msg = msg + "text: " + str(response.text)
    msg = msg + "-*-" * 20 + " " * 4 + method + " End " + " " * 4 + "-*-" * 20 + "\n"

    with open(log_file, 'a') as f:
        f.write(msg)


@click.group()
def cli():
    pass


@click.command()
def snapshot():
    '''
    备份配置文件中指定的索引
    '''
    es_handle = es_handler()
    max_snapshot_bytes_per_sec = es_conf["max_snapshot_bytes_per_sec"]
    max_restore_bytes_per_sec = es_conf['max_restore_bytes_per_sec']
    indexs = parse_index(es_conf['indexs']['snapshot'])
    for delay_day, indices in indexs.items():
        date_fmt = convert_date(delay_day, '%Y%m%d')
        es_handle.put_repository(date_fmt, max_snapshot_bytes_per_sec, max_restore_bytes_per_sec)
        es_handle.put_snapshot(indices, date_fmt)


@click.command()
def delete():
    '''
    删除配置文件中指定的索引
    '''
    es_handle = es_handler()
    indexs = parse_index(es_conf['indexs']['delete'])
    indice_list = []
    for value in indexs.values():
        indice_list += value
    indices = ','.join(list(set(indice_list)))
    es_handle.delete_repository(indices)


@click.command()
def get_snapshot():
    '''
    获取备份索引信息
    '''
    es_handle = es_handler()
    indexs = parse_index(es_conf['indexs']['snapshot'])
    for delay_day in indexs.keys():
        date_fmt = convert_date(delay_day, '%Y%m%d')
        response = es_handle.get_snapshot(date_fmt).json()
        print json.dumps(response, indent=4)


@click.command()
def del_snapshot():
    '''
    删除指定日期备份
    '''
    es_handle = es_handler()
    indexs = parse_index(es_conf['indexs']['delete'])
    for delay_day in indexs.keys():
        date_fmt = convert_date(delay_day, '%Y%m%d')
        response = es_handle.delete_snapshot(date_fmt).json()
        print json.dumps(response, indent=4)


@click.command()
def put_tag():
    '''给索引设置标签'''
    es_handle = es_handler()
    indexs = es_conf['indexs']['setting']
    for index in indexs:
        if index['enable'] != 0:
            fmt = _formats.get(index['date_fmt'], "%Y.%m.%d")
            date_fmt = convert_date(index['delay_days'], fmt)
            indices = ["%s%s" % (item, date_fmt) for item in index['index']]
            es_handle.put_tag(indices, index['tag'])


def main():
    cli.add_command(snapshot)
    cli.add_command(delete)
    cli.add_command(get_snapshot)
    cli.add_command(del_snapshot)
    cli.add_command(put_tag)
    cli()


if __name__ == '__main__':
    try:
        main()
    except Exception as e:
        print e
