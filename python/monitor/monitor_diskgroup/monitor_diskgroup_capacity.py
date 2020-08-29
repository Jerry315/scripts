#!/usr/bin/env python
# -*- coding: utf-8 -*-
'''
运行环境：python2.7，需要安装requests模块

'''
import requests
import click
import json
from common import logger, get_group_id, get_data, config


def get_group_capacity(gid, url, percent):
    used = 0
    total = 0
    data = get_data(url)
    if data:
        for item in data['monitor_info']:
            if item['group_id'] == gid:
                for disk in item['disk_infos']:
                    used += disk['used']
                    total += disk['size']
    if percent:
        return '%.2f' % ((float(total - used) / float(total)) * 100,)
    return (total - used)


def get_zero_group_capacity(disk_type=2, gid=0):
    import time
    zero_data = {}
    query_data = {"data": []}
    data = requests.get(config['codisktracker']['group_info_url'],
                        params={"disk_type": disk_type, "group_id": gid, "time": int(time.time()) - 86400})
    if data.status_code == 200:
        data = data.json()["group_infos"][0]["disc_infos"]
        for item in data:
            if item["is_online"] == 0:
                continue
            if item["cycle"] not in zero_data:
                zero_data[item["cycle"]] = {}
            total_capacity = (item["used"] + item["left"]) / (1024 * 1024 * 1024 * 1024)
            if 0 < total_capacity < 2:
                capacity = 2
            elif 2 < total_capacity < 4:
                capacity = 4
            else:
                capacity = 8
            if capacity not in zero_data[item["cycle"]]:
                zero_data[item["cycle"]][capacity] = {"count": 0, "left": 0}
            zero_data[item["cycle"]][capacity]["count"] += 1
            zero_data[item["cycle"]][capacity]["left"] += item["left"]

        for cycle, disk in zero_data.items():
            for space in disk:
                query_data["data"].append({"{#STORAGE_CYCLE}": str(cycle), "{#DISK_SPACE}": str(space)})
    return zero_data, query_data


@click.group()
def cli():
    pass


@click.command()
@click.option("-m", "--method")
@click.option("-g", "--gid", type=int, required=False)
@click.option("-p", "--percent", default=False, required=False)
def normal(method, gid, percent):
    url = config['codisktracker']['mointor_url']
    if method == 'group':
        result = get_group_id(url, get_data)
    elif method == 'capacity':
        result = get_group_capacity(gid, url, percent)
    else:
        result = 0
    print result


@click.command()
@click.option("-c", "--cycle", type=int, required=False)
@click.option("-s", "--space", type=int, required=False)
@click.option("-a", "--arg", type=str, required=False)
def zero(arg, cycle, space):
    if cycle == 0 or cycle and space and arg:
        zero_data, _ = get_zero_group_capacity(disk_type=2, gid=0)
        print zero_data[cycle][space][arg]
    else:
        _, query_data = get_zero_group_capacity(disk_type=2, gid=0)
        print json.dumps(query_data)


if __name__ == '__main__':
    try:
        cli.add_command(normal)
        cli.add_command(zero)
        cli()
    except Exception as e:
        logger.log(str(e), mode=False)
        print 0
