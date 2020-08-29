#!/usr/bin/env python
# -*- coding: utf-8 -*-
import requests, os

center_inner_url = 'http://api.nginx.private/internal/devices/token?cid='
center_api_url = 'https://jxsr-api.antelopecloud.cn'
# center_api_url = 'https://dgdx-api.antelopecloud.cn/'


class SyncDeviceName(object):
    def __init__(self, center_api_url, center_inner_url, headers):
        self.center_api_url = center_api_url
        self.center_inner_url = center_inner_url
        self.headers = headers

    def get_token(self, cids):
        token_list = []
        for cid in cids:
            url = self.center_inner_url + str(cid)
            headers = {
                'Content-Type': 'application/json'
            }
            response = requests.get(url, headers=headers, timeout=30).json()
            token_list.append({'cid': cid, 'token': response['token']})
        return token_list

    def bind_device(self, bind_info):
        cids = bind_info.keys()
        token_list = self.get_token(cids)
        if token_list:
            for token in token_list:
                url = self.center_api_url + '/cloudeye/v1/devices/{cid}/bindmapping'.format(cid=token['cid'])
                querystring = {"client_token": token['token']}
                payload = "{\n  \"action\": \"bind\",\n  \"sn\": \"%s\"\n}" % bind_info[token['cid']]
                response = requests.request('POST', url, data=payload, headers=self.headers, params=querystring)
                print response.status_code
                print response.text
        else:
            print 'cids.2018-12-19.2018-12-18 %s has no token' % cids


if __name__ == '__main__':
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))
    cid_file = os.path.join(BASE_DIR, 'cid.txt')
    headers = {
        'Content-Type': "application/json",
        'Cache-Control': "no-cache"
    }
    if not os.path.exists(cid_file):
        exit('cid.txt file is not exist!')
    try:
        bind_info = {}
        with open(cid_file, 'rb') as f:
            for line in f:
                bind_info[int(line.split()[1])] = int(line.split()[0])
        sdn = SyncDeviceName(center_api_url, center_inner_url, headers)
        sdn.bind_device(bind_info)
    except Exception as e:
        print e
