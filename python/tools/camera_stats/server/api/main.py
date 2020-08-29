#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json
import requests
from flask import Blueprint, request, jsonify
from utils.tk import Token
from settings import config
from utils.mongo_handle import DeviceModel

bp = Blueprint("camera", __name__, url_prefix="/camera/v1")
qtx = Token(config["username"])


@bp.route("/token", methods=["POST"])
def token():
    response = {
        "code": 0,
        "msg": "登录成功",
        "token": None
    }
    username = request.json.get("username")
    password = request.json.get("password")
    if username == config["username"] and password == config["password"]:
        response["token"] = qtx.generate_auth_token()
    else:
        response["msg"] = "用户名或密码错"
        response["code"] = 40001
    return jsonify(response)


@bp.route("/records", methods=["GET"])
def records():
    cid_info = dict()
    result = []
    response = {
        "code": 0,
        "msg": "获取数据成功",
        "data": None
    }
    token = request.args.get("token")
    size = int(request.args.get("size", 100))
    page = int(request.args.get("page", 1))
    start = int(request.args.get("start"))
    end = int(request.args.get("end"))
    cids = request.args.get("cids")
    if cids:
        cids = json.loads(cids)
        if len(cids) > 10000:
            response["code"] = 40002
            response["msg"] = "cid 列表超过10000"
            return jsonify(response)
    if int(((end - start) / 86400)) > 7:
        response["code"] = 40003
        response["msg"] = "查询时间轴超过7天"
        return jsonify(response)
    if qtx.verify_auth_token(token):
        if cids:
            filter = {"cid": {"$in": cids}, "create_time": {"$gte": start, "$lte": end}}
        else:
            filter = {"create_time": {"$gte": start, "$lte": end}}
        dm = DeviceModel()
        data = dm.find(filter, limit=size, skip=(page - 1) * size)
        if data:
            device_info_request = requests.get(config['device_info_url'], timeout=10)
            if device_info_request.status_code == 200:
                device_info = device_info_request.json()
            else:
                device_info = dict()

            for record in data:
                record.pop("_id")
                record.pop("timer")
                record.pop("tm_hour")
                record["group"] = None
                record["name"] = None
                record["sn"] = None
                record["brand"] = None
                record["model"] = None
                if record["cid"] in cid_info:
                    record["group"] = cid_info[record["cid"]]["group"]
                    record["name"] = cid_info[record["cid"]]["name"]
                    record["sn"] = cid_info[record["cid"]]["sn"]
                    record["brand"] = cid_info[record["cid"]]["brand"]
                    record["model"] = cid_info[record["cid"]]["model"]
                else:
                    for item in device_info["data"]:
                        if record["cid"] == item["_id"]:
                            cid_info[record["cid"]] = {
                                "name": item["name"],
                                "group": item["group"],
                                "sn": item["sn"],
                                "brand": item["brand"],
                                "model": item["model"]
                            }
                            record["group"] = item["group"]
                            record["name"] = item["name"]
                            record["sn"] = item["sn"]
                            record["brand"] = item["brand"]
                            record["model"] = item["model"]
                            break
                result.append(record)
            response["data"] = result
        else:
            response["code"] = 40004
            response["msg"] = "参数不合法"
    else:
        response["code"] = 40001
        response["msg"] = "token失效"
    return jsonify(response)
