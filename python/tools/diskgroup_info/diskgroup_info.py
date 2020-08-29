#!/usr/bin/env python
# -*- coding: utf-8 -*-
'''
运行环境：python2.7，需要安装requests模块
在运行之前可以传入要查看的group_id，只显示指定group信息
url = 'http://codisktracker.private:8201/console/group_info'
params = {'disk_type':2}
'''
import json
import optparse
import os
import sys
import requests
import yaml
import time
from datetime import datetime
from operator import itemgetter
from requests.packages.urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)


def get_data(url, params, gid=None, display=False, period=None, detail=False):
    '''
    :param url: 磁盘调度HTTP接口
    :param params: disk_type,磁盘类型 1：循环存储磁盘 2：循环对象磁盘 3：永久对象磁盘
    :param gid: 磁盘组id
    :return:
    '''
    url = url + '/console/group_info'
    result = requests.get(url, params)
    context = result.text.encode('utf-8').decode('utf-8')
    if context:
        data = json.loads(context)
        group_infos = data.get('group_infos')
        if group_infos:
            parse_data(group_infos, gid, display, period, detail)


def get_first_time():
    tmp = {}
    url = config['codisktracker']['codisktracker_url'] + '/console/get_disc_first_time'
    disc_first_time_infos = requests.get(url).json()['disc_first_time_infos']
    for item in disc_first_time_infos:
        tmp[item['disc_id']] = item['first_time']
    return tmp


def parse_data(data, gid=None, display=False, period=None, detail=False):
    cycle_dict = dict()
    capacity_dict = dict()
    disc_first_time = get_first_time()
    for item in data:
        group_id = item.get('group_id')
        group_type = item.get('group_type')
        onlines = item.get('onlines')
        create_time = datetime.fromtimestamp(item.get('create_time')).strftime('%Y-%m-%d %H:%M:%S')
        cycle = item.get('cycle')
        if gid or gid == 0:
            if group_id != gid:
                continue
        if period and cycle != period:
            continue
        if gid == None:
            if 'cycle_%d' % cycle in cycle_dict:
                cycle_dict['cycle_%d' % cycle] += 1
            else:
                cycle_dict['cycle_%d' % cycle] = 1
        lockTime = item.get('lockTime')
        dispatcher_id = item.get('dispatcher_ids')
        if dispatcher_id:
            dispatcher_id = dispatcher_id[0]
        disc_infos = item.get('disc_infos')
        ip_dict = {}
        total_left = 0
        total_used = 0
        disk_list = []
        for disk in disc_infos:
            left = int(disk.get('left'))
            used = int(disk.get('used'))
            total_used += used
            total_left += left
            if gid == 0:
                capacity = left + used
                if (capacity / (1024 * 1024 * 1024 * 1024)) > 4:
                    if 8 in capacity_dict:
                        capacity_dict[8] += 1
                    else:
                        capacity_dict[8] = 1
                elif 4 >= (capacity / (1024 * 1024 * 1024 * 1024)) > 2:
                    if 4 in capacity_dict:
                        capacity_dict[4] += 1
                    else:
                        capacity_dict[4] = 1
                elif 2 >= (capacity / (1024 * 1024 * 1024 * 1024)) > 1:
                    if 2 in capacity_dict:
                        capacity_dict[2] += 1
                    else:
                        capacity_dict[2] = 1
                else:
                    if 1 in capacity_dict:
                        capacity_dict[1] += 1
                    else:
                        capacity_dict[1] = 1
            if detail:
                discID = disk.get('discID')
                first_time = disc_first_time.get(discID, 0)
                public_ip = disk.get('public_ip')
                local_ip = disk.get('local_ip')
                if local_ip in ip_dict:
                    ip_dict[local_ip] = ip_dict[local_ip] + 1
                else:
                    ip_dict[local_ip] = 1
                port = disk.get('port')
                is_online = disk.get('is_online')
                if (left / float(1024 * 1024 * 1024)) > 1024:
                    left = str('%.2f' % (left / float(1024 * 1024 * 1024 * 1024),)) + 'TB'
                else:
                    left = str('%.2f' % (left / float(1024 * 1024 * 1024))) + 'GB'
                disk_list.append({
                    'first_time': first_time,
                    'discID': discID,
                    'is_online': is_online,
                    'public_ip': public_ip,
                    'local_ip': local_ip,
                    'port': port,
                    'left': left
                })
        if (total_used / float(1024 * 1024 * 1024)) > 1024:
            total_used = str('%.2f' % (total_used / float(1024 * 1024 * 1024 * 1024))) + 'TB'
        else:
            total_used = str('%.2f' % (total_used / float(1024 * 1024 * 1024))) + 'GB'
        if (total_left / float(1024 * 1024 * 1024)) > 1024:
            total_left = str('%.2f' % (total_left / float(1024 * 1024 * 1024 * 1024))) + 'TB'
        else:
            total_left = str('%.2f' % (total_left / float(1024 * 1024 * 1024))) + 'GB'
        tmp_group = "[%s] [Getdiskinfo; group_id: %s, group_type: %s, cycle:%s, onlines: %s, lockTime: %s, dispatcher_id: %s, total_used: %s, total_left: %s ]" % (
            create_time, group_id, group_type, cycle, onlines, lockTime, dispatcher_id, total_used, total_left)
        print tmp_group
        disk_list = sorted(disk_list, key=itemgetter('first_time'))
        tmp_list = []
        seq = 0
        for item in disk_list:
            if item['first_time']:
                item['first_time'] = datetime.fromtimestamp(item['first_time']).strftime('%Y-%m-%d %H:%M:%S')
            disk_str = "\t[first_time: %s, seq: %d, discID: %s,  is_online: %s, public_ip: %s, local_ip: %s, port: %s, left: %s]" % (
                item['first_time'], seq, item['discID'], item['is_online'], item['public_ip'], item['local_ip'],
                item['port'], item['left'])
            tmp_list.append(disk_str)
            seq += 1
        if tmp_list:
            print '\n'.join(tmp_list)
        for k, v in ip_dict.items():
            if v == 1:
                ip_dict.pop(k)
        if ip_dict and display:
            print '\tGroup %s 重复的IP信息：' % group_id
            print '\t' + '\n\t'.join(['ip: %s, repeat_count: %s' % (k, v) for k, v in ip_dict.items()])
    if cycle_dict:
        print '[' + '%s' % datetime.today().strftime('%Y-%m-%d %H:%M:%S') + '] ' + '[' + ', '.join(
            ["%s: %s" % (c, d) for c, d in cycle_dict.items()]
        ) + ']'
    if capacity_dict:
        print '[' + '%s' % datetime.today().strftime('%Y-%m-%d %H:%M:%S') + '] ' + '[' + ', '.join(
            ["disk %dTB count: %d" % (x, y) for x, y in capacity_dict.items()]) + ']'


if __name__ == '__main__':
    try:
        BASE_DIR = os.path.dirname(os.path.abspath(__file__))
        yaml_file = os.path.join(BASE_DIR, 'config.yml')
        config = yaml.load(open(yaml_file, 'rb'))['config']
        limit = config['limit']
        usage = "python diskgroup_info.py [-t<disk_type>] [-g<group_id>] [arg1,...] -d -a\n parameter 'disk_type' must pass"
        parser = optparse.OptionParser(usage)
        parser.add_option('-a', '--appear', action="store_false", dest="appear", help="view repeat ip info")
        parser.add_option('-c', '--cycle', dest='cycle', default=None, help='storate cycle 0 or 7 or 30')
        parser.add_option('-d', '--detail', action="store_false", dest='detail', help='view disk detail info')
        parser.add_option('-u', '--url', dest='url', default=None, help='codisktracker_url')
        parser.add_option('-t', '--disk_type', dest='disk_type', default=2,
                          help='disk type 1: cycle storage disk 2: cycle object disk 3: Permanent object disk, defaut disk_type is 2')
        parser.add_option('-g', '--gid', dest='group_id', default=None, help='disk group id')
        argv = sys.argv[1:]
        if '-a' in argv:
            display = True
        else:
            display = False
        if '-d' in argv:
            detail = True
        else:
            detail = False
        (options, args) = parser.parse_args(argv)
        url = options.url
        if not url:
            url = config['codisktracker']['codisktracker_url']
        disk_type = options.disk_type
        try:
            gid = int(options.group_id)
        except Exception:
            gid = None
        try:
            cycle = int(options.cycle)
        except Exception:
            cycle = None
        if url and disk_type:
            get_data(url, {'disk_type': disk_type, "time": int(time.time()) - limit}, gid, display, cycle, detail)
        else:
            parser.print_help()
    except Exception as e:
        print e
