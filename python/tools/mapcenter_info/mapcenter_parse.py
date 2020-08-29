# -*- coding: utf-8 -*-
import requests
import sys
import optparse
from common import Logger, SendMail, config
from prettytable import PrettyTable

cycle_info = {
    1: 7,
    2: 30,
    3: 90,
    4: 15
}


def parse(url):
    result = requests.get(url, timeout=5)
    if result.status_code == 200:
        data = result.json()
        mapinfo = data['mapinfo']
        for i in range(len(mapinfo)):
            for j in range(len(mapinfo) - i - 1):
                if cycle_info[mapinfo[j]["CoDispatcherMark"]["Cycle"]] > cycle_info[
                    mapinfo[j + 1]["CoDispatcherMark"]["Cycle"]]:
                    mapinfo[j], mapinfo[j + 1] = mapinfo[j + 1], mapinfo[j]
        return data
    else:
        print result.content


def search(url, peerid=None, ip=None, port=None, cycle=None, alarm=False):
    data = parse(url)
    result = []
    if data:
        AlarmStatus = data['AlarmStatus']
        mapinfo = data['mapinfo']
        if alarm:
            print AlarmStatus
        for item in mapinfo:
            tmp = dict()
            tmp['indexs'] = []
            tmp['PeerID'] = item["CoDispatcherMark"]["PeerID"]
            tmp['Cycle'] = item["CoDispatcherMark"]["Cycle"]
            if peerid and item["CoDispatcherMark"]["PeerID"] != peerid:
                continue
            if cycle and item["CoDispatcherMark"]["Cycle"] != cycle:
                continue
            if ip and port:
                for index in item['Indexs']:
                    if index['ip'] == ip and index['rpc_port'].strip(":") == port:
                        tmp['indexs'].append(index)
            elif ip:
                for index in item['Indexs']:
                    if index['ip'] == ip:
                        tmp['indexs'].append(index)
            elif port:
                for index in item['Indexs']:
                    if index['rpc_port'].strip(":") == port:
                        tmp['indexs'].append(index)
            else:
                tmp['indexs'] = item['Indexs']
            if tmp['indexs']:
                result.append(tmp)
    return result


def table(data):
    t = PrettyTable(["PeerID", "Cycle", "Index", "IndexCount"])
    t.align = 'c'
    t.valign = 'm'
    for item in data:
        tmp = []
        item_cycle = item['Cycle']
        for k, v in cycle_info.items():
            if k == item['Cycle']:
                item_cycle = v
        for index in item['indexs']:
            tmp.append(index['ip'] + index['rpc_port'])
        t.add_row([item['PeerID'], item_cycle, '\n'.join(tmp), len(item['indexs'])])
        t.add_row(['-' * len(str(item['PeerID'])), '-' * len('Cycle'), '-' * len(tmp[0]), '-' * len("IndexCount")])
    print t


if __name__ == '__main__':
    try:
        usage = "./start.sh [-c<cycle>] [-p<peerid>] [-i<indexIP>] [-p<indexPort>] [-a]"
        parser = optparse.OptionParser(usage)
        parser.add_option('-a', '--alarm', dest='alarm', action="store_false", help='display alarm message')
        parser.add_option('-c', '--cycle', dest='cycle', help='1 map 7，2 map 30，3 map 90，4 map 15')
        parser.add_option('-i', '--ip', dest='ip', help='indexIP')
        parser.add_option('-p', '--peerid', dest='peerid', help='PeerID')
        parser.add_option('-P', '--port', dest='port', help='indexPort')
        argv = sys.argv[1:]
        if '-a' in argv:
            argv.remove('-a')
            alarm = True
        else:
            alarm = False
        (options, args) = parser.parse_args(argv)
        peerid = options.peerid
        if peerid:
            peerid = int(peerid)
        cycle = options.cycle
        if cycle:
            cycle = int(cycle)
        ip = options.ip
        port = options.port
        url = config['url']
        data = search(url, peerid, ip, port, cycle, alarm)
        table(data)
    except Exception as e:
        print str(e)
