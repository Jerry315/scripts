# -*- coding: utf-8 -*-
import json
import time
import requests
from datetime import datetime
from threadpool import makeRequests
from common import Logger
from mongo_handle import DeviceModel
from settings import pool, cid_files

logger = Logger()
dm = DeviceModel()


def get_cids(url, current_time, timeout):
    """
    获取cid信息，返回两种类型的cid列表，一种是正常cid，一种是心跳超时60秒以上的cid
    :param url: cid_info_url
    :param current_time: 当前时间戳
    :param timeout: 心跳超时时间
    :return:
    """
    try:
        data = requests.get(url, timeout=10)
        if data.status_code == 200:
            logger.info("[get_cids] get data success")
            data = data.json()
        else:
            raise Exception(data.content)
        # 存放正常的cid
        normal_cid_list = []

        # 存放心跳超时的cid
        unnormal_cid_list = []
        for cid_info in data['data']:
            message_timestamp = cid_info.get("message_timestamp", 0)
            if current_time - message_timestamp < timeout:
                if cid_info.get("push_state") == 4:
                    normal_cid_list.append(cid_info["_id"])
                else:
                    unnormal_cid_list.append(cid_info["_id"])
            else:
                unnormal_cid_list.append(cid_info["_id"])
        return (normal_cid_list, unnormal_cid_list)
    except Exception as e:
        logger.error(str(e))


def get_device_info(url):
    """
    获取设备信息，静态字段，name，group，sn等
    :param url:
    :return:
    """
    try:
        data = requests.get(url, timeout=10)
        if data.status_code == 200:
            logger.info("[get_device_info] get data success")
            data = data.json()
        else:
            raise Exception(data.content)
        return data
    except Exception as e:
        logger.error(str(e))


def db_insert(doc, current_time, tm_hour, t1):
    """
    插入数据，插入前对数据进行验证知否存在，已存在就忽略
    :param doc:
    :param current_time:
    :param tm_hour:
    :param t1:
    :return:
    """
    tmp = dict()
    doc['cid'] = doc.pop('_id')
    doc['tm_hour'] = tm_hour
    data = dm.find({"cid": doc['cid'], "tm_hour": tm_hour, "create_time": {"$gte": t1, "$lte": t1 + 3659}})
    if not [item for item in data]:
        tmp['_id'] = dm.next_id()
        tmp['cid'] = doc['cid']
        tmp['tm_hour'] = tm_hour
        tmp['create_time'] = current_time
        tmp['status'] = doc['status']
        tmp['timer'] = datetime.utcnow()
        dm.insert(tmp)


def gather_data(cid_info_url, device_info_url, timeout=60):
    """
    数据校对，写入数据库
    :param cid_info_url:
    :param device_info_url:
    :param timeout:
    :return:
    """
    current_time = int(time.time())
    t1 = time.mktime(time.strptime(
        datetime.fromtimestamp(time.time()).replace(minute=0, second=0).strftime(
            '%Y-%m-%d %H:%M:%S'), '%Y-%m-%d %H:%M:%S'))
    normal_cid_list, unnormal_cid_list = get_cids(cid_info_url, int(round(time.time() * 1000)), timeout)
    device_data = get_device_info(device_info_url)
    device_list = []
    func_var = []
    tm_hour = time.localtime(current_time).tm_hour
    for device_info in device_data['data']:

        # 根据cid在不同的cid列表中，判断其状态
        if device_info["_id"] in normal_cid_list:
            device_info['status'] = "Y"
        elif device_info["_id"] in unnormal_cid_list:
            device_info['status'] = "N"
        else:
            continue
        device_list.append(device_info)
        func_var.append(([device_info, current_time, tm_hour, t1], None))
    requests = makeRequests(db_insert, func_var)
    [pool.putRequest(req) for req in requests]
    pool.wait()

    with open('%s.%s' % (cid_files, datetime.today().strftime("%Y-%m-%d")), 'w') as f:
        json.dump({"data": device_list}, f)


def get_period_data(cid_list, start_time, end_time):
    logger.info("start get data from mongodb")
    filter = {"cid": {"$in": cid_list}, "create_time": {"$gte": start_time, "$lte": end_time}}
    dm = DeviceModel()
    records = dm.find(filter)
    logger.info("get data from mongodb finish")
    return [record for record in records]


if __name__ == '__main__':
    pass
