#!/usr/bin/env python
# -*- coding: utf-8 -*-
import requests
import json
import time
from settings import config
from logger import Logger


def get_token(url):
    cig_url = url + "/cigweb/v1/server/login"
    payload = {"app_id": "fxqc", "app_secret": "zix6u1596r7ae28dle12b30hh5qc142e"}
    headers = {
        'Content-Type': "application/json"
    }
    response = requests.request("POST", cig_url, data=json.dumps(payload), headers=headers).json()
    return response["data"]["access_token"]


def parse_data(url, token):
    cig_url = url + "/cigweb/v1/server/cigs"
    result = dict()
    payload = {"page": 1, "size": 200, "sort": "first_online", "desc": 1, "mode": 1}
    headers = {
        'Content-Type': "application/json",
        'Authorization': token,
    }
    response = requests.request("POST", cig_url, data=json.dumps(payload), headers=headers).json()
    current_time = int(time.time())
    result["total"] = len(response["data"]["cigs"])
    result["online"] = 0
    result["offline"] = 0
    result["data"] = []
    result["time"] = current_time
    model_info = dict()
    for item in response["data"]["cigs"]:
        if item["status"] == 1:
            result["online"] += 1
        else:
            result["offline"] += 1
        if item["model"] + ":" + item["software"] not in model_info:
            model_info[item["model"] + ":" + item["software"]] = 0
        model_info[item["model"] + ":" + item["software"]] += 1

    for m, c in model_info.items():
        tmp = dict()
        model, software = m.split(":")
        tmp["model"] = model
        tmp["software"] = software
        tmp["count"] = c
        tmp["time"] = current_time
        result["data"].append(tmp)
        logger.log(json.dumps(tmp))
    logger.log(json.dumps(result))


if __name__ == '__main__':
    logger = Logger(config['cigweb']['access_log'], config['cigweb']['error_log'])
    token = get_token(config["cigweb"]["url"])
    parse_data(config["cigweb"]["url"], token)
