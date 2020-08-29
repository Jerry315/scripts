# -*- coding: utf-8 -*-
"""
   批量重启设备
"""
import sys
import json
import threading
import requests


login_path = '{}/tong/v1/users/login'
reboot_path = '{}/tong/v1/camera/{}/restart'


def login(host, account, password):
    """登录通配后台并获取 jwtoken"""
    headers = {
        'Content-Type': 'application/json',
    }
    payload = {
        'account': account,
        'password': password,
    }
    url = login_path.format(host)
    resp = requests.request('POST', url, data=json.dumps(payload), headers=headers)
    if resp.status_code != 200:
        print u'用户登录失败'
        print resp.text
        sys.exit(-1)
    jwtoken = resp.json().get('data', '').get('token')
    if not jwtoken:
        print u'登录没有获取到 jwtoken'
        print resp.text
        sys.exit(-1)
    return jwtoken


def reboot(cid, host, token):
    """重启设备"""
    headers = {
        'Authorization': "Bearer %s" % token,
        'Content-Type': 'application/json',
    }
    payload = ""
    url = reboot_path.format(host, cid)
    response = requests.request("POST", url, data=payload, headers=headers)
    if response.status_code != 200:
        print "重启%s失败" % cid
        print response.text
    else:
        print response.text


if __name__ == '__main__':
    """
    host：中央api地址
    account：通配的管理员账号
    password：通配的管理员密码
    """
    host = "https://dgdx-api.antelopecloud.cn"
    account = "xxx@topvdn.com"
    password = "xxxx"
    token = login(host, account, password)
    with open('cids_example.txt','r') as f:
        cids = []
        for cid in f:
            cids.append(int(cid))
    if token:
        threads = []
        for cid in cids:
            t = threading.Thread(target=reboot, args=(cid, host, token))
            threads.append(t)

        for t in threads:
            t.start()

        for t in threads:
            t.join()
