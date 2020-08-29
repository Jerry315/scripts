# -*- coding: utf-8 -*-
import yaml
import os
import requests
import json
import time
from prettytable import PrettyTable
from common import BASE_DIR


class CheckDevicePeriod(object):
    def __init__(self, url):
        self.url = url
        self.response = {}

    def check(self, checkmode, check_cycle, cids, retry_count=3):
        querystring = {"checkmode": checkmode, "check_cycle": check_cycle}
        payload = "------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"cidlist\"\r\n\r\n{\"cidlist\": %s}\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--" % cids
        headers = {
            'content-type': "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
        }
        while retry_count > 0:
            result = requests.post(self.url, data=payload, headers=headers, params=querystring)
            if result.status_code == 200:
                data = result.json()
                if data['cids.2018-12-19.2018-12-18']:
                    return data['cids.2018-12-19.2018-12-18']
                else:
                    break
            else:
                time.sleep(12)
                result = requests.post(self.url, data=payload, headers=headers, params=querystring)
                if result.status_code == 200:
                    data = result.json()
                    if data['cids.2018-12-19.2018-12-18']:
                        return data['cids.2018-12-19.2018-12-18']
                    else:
                        break
                self.response['code'] = result.status_code
                self.response['msg'] = result.content
                # logger.log(json.dumps(self.response), mode=False)
            retry_count -= 1

    def report(self):
        checkmodes = config['checkmodes']
        result = []
        for checkmode in checkmodes:
            for area in config['cids_info']:
                cids = config['cids_info'][area]['cids.2018-12-19.2018-12-18']
                if checkmode == 0:
                    cycle = 2
                else:
                    cycle = config['cids_info'][area]['cycle']
                tmp = {}
                tmp['area'] = config['cids_info'][area]['name']
                tmp['cycle'] = cycle
                tmp['checkmode'] = checkmode
                tmp['data'] = []
                cid_len = len(cids)
                index = 0
                while True:
                    cidlist = cids[index:index + step]
                    data = self.check(checkmode, cycle, cidlist)
                    if data:
                        tmp['data'] = tmp['data'] + data
                    if cid_len <= step:
                        break
                    if (cid_len - index) < step and (cid_len - index) > 0:
                        index += (cid_len - index)
                    elif cid_len - index > step:
                        index += step
                    else:
                        break
                    time.sleep(10)
                result.append(tmp)
        t = PrettyTable(["Area", "CheckMode", "CheckCycle", "RealCycle", "CID"])
        t.align = 'c'
        t.valign = 'm'
        for item in result:
            cid_tmp = []
            cycle_tmp = []
            for info in item['data']:
                cid_tmp.append(str(info['CID']))
                cycle_tmp.append(str(info['Cycle']))
                t.add_row([item['area'], item['checkmode'], item['cycle'], info['Cycle'], info['CID']])
            if cid_tmp:
                t.add_row([item['area'], item['checkmode'], item['cycle'], '\n'.join(cycle_tmp), '\n'.join(cid_tmp)])
                t.add_row(['-' * (len(item['area']) + 2), '-' * (len('CheckMode') + 2), '-' * (len("CheckCycle") + 2),
                           '-' * (len("RealCycle") + 2), '-' * (len(str(info['Cycle'])) + 2)])
        print t
        # return str(t)


if __name__ == '__main__':
    # logger = Logger()
    yaml_file = os.path.join(BASE_DIR, 'check_store_period.yml')
    config = yaml.load(open(yaml_file, 'rb'))['config']
    # try:
    url = config['url']
    step = config['step']
    cdp = CheckDevicePeriod(url)
    cdp.report()
    # receivers = config['smtp']['receivers']
    # subject = config['subject']
    # SendMail().send(receivers, subject, msg)
    # except Exception as e:
    #     logger.log(str(e), mode=False)
