#!/usr/bin/env python
# -*- coding: utf-8 -*-
import requests
import json
import sys
import os
import time
import datetime
from collections import OrderedDict
from lib.settings import (
    wechat_corpid,
    token_url,
    send_url,
    time_format,
    alert_url,
    record_file_dir,
    alert_info
)
from lib.logger import Logger
from lib.mail import SendMail
from lib.redis_handler import rd


def convert_seconds_to_time(second):
    a = "{:0>8}".format(datetime.timedelta(seconds=second))
    return a


def _decode_list(data):
    rv = []
    for item in data:
        if isinstance(item, unicode):
            item = item.encode('utf-8')
        elif isinstance(item, list):
            item = _decode_list(item)
        elif isinstance(item, dict):
            item = _decode_dict(item)
        rv.append(item)
    return rv


def _decode_dict(data):
    rv = {}
    for key, value in data.iteritems():
        if isinstance(key, unicode):
            key = key.encode('utf-8')
        if isinstance(value, unicode):
            value = value.encode('utf-8')
        elif isinstance(value, list):
            value = _decode_list(value)
        elif isinstance(value, dict):
            value = _decode_dict(value)
        rv[key] = value
    return rv


def get_string_between_list_two_keys(first_key, second_key, the_list, default_choice=1, sep_str=" "):
    '''
        可以获取到list中的一段数据，可以根据2个关键字来截取
    '''
    _ret_list_2 = []
    the_list = _decode_list(the_list)
    begin = 0
    all_string = ""
    for love in range(len(the_list)):
        if str(the_list[love]).startswith(first_key):
            begin = love + 1
            continue
        if str(the_list[love]).startswith(second_key):
            if begin == 0:
                continue
            end = love
            _ret_list_2.append((begin, end))
            begin = 0
    if default_choice == "all":
        for ak in _ret_list_2:
            t_1 = sep_str.join([str(ok) for ok in the_list[ak[0]:ak[1]]])
            all_string += t_1
            all_string += sep_str
    else:
        for i in range(int(default_choice)):
            first = _ret_list_2[i][0]
            second = _ret_list_2[i][1]
            t_2 = sep_str.join([str(ok) for ok in the_list[first:second]])
            all_string += t_2
            all_string += sep_str
    return all_string.lstrip(sep_str).rstrip(sep_str)


def insert_into_list(ori_list, match_string, insert_data):
    jk = ""
    for i in range(len(ori_list)):
        if str(ori_list[i]).startswith(str(match_string)):
            jk = i
            break
    if jk != "":
        ori_list.insert(jk, insert_data)
    return ori_list


def get_token(wechat_corpsecret, agentid):
    """
    token保存在redis中，过期时间2小时
    :param wechat_corpsecret:
    :param agentid:
    :return:
    """
    params = {
        "corpid": wechat_corpid,
        "corpsecret": wechat_corpsecret
    }
    response = requests.get(token_url, params=params, timeout=15)
    if response.status_code == 200:
        access_token = response.json()['access_token']
        # redis version under 3.0
        #rd.setex('%s_token' % agentid, access_token, 7200)
        # redis version 3.2
        try:
            rd.setex('%s_token' % agentid, 7200, access_token)
        except Exception:
            rd.setex('%s_token' % agentid, access_token, 7200)
        return access_token
    else:
        logger.error("get token failed")


def p1_record(eventid, hostname, trigger, P, flag=True):
    """
    记录未回复的告警，删除已恢复的
    :param eventid: 告警事件id
    :param hostname: 告警主机
    :param trigger: 触发器内容
    :param flag: 如果是True表示告警已恢复，可以删除，如果是False，新增未恢复告警
    :return:
    """
    records = OrderedDict()
    record_file = os.path.join(record_file_dir, "%s.txt" % P)
    # 检测保存未恢复告警文件是否存在，不存在则创建
    if not os.path.exists(record_file):
        if not os.path.exists(os.path.dirname(record_file)):
            os.makedirs(os.path.dirname(record_file))
        with open(record_file, 'w') as f:
            pass

    if not os.path.getsize(record_file):
        data = {}
    else:
        with open(record_file, 'r') as r:
            data = json.load(r)
    if flag:
        n = 1
        if data:
            for k, v in data.items():
                if v['event_id'] == eventid:
                    data.pop(k)
                    continue
                else:
                    records[n] = v
                n += 1
    else:
        n = 1
        tmp = {
            "event_id": eventid,
            "host": hostname,
            "trigger": trigger
        }
        if data:
            for k, v in data.items():
                records[n] = v
                n += 1
        records[n] = tmp
    with open(record_file, 'w') as f:
        json.dump(records, f, indent=4)


def send_wechat(msg, level):
    """
    微信消息接口限制1分钟内给单个ID发送消息不超过30条；
    timer计时，timer失效时间一分钟，一分钟内counter不能超过30，超过30需要等待，timer失效，counter计数清0
    :param msg:
    :param level:
    :return:
    """
    log_msg = {
        "code": None,
        "msg": None,

    }
    s = alert_info.get(level, None)
    if not s:
        s = alert_info.get("Other")
    agentid = s['agentid']
    wechat_corpsecret = s['wechat_corpsecret']
    toparty = s["toparty"]
    # 获取token，优先从redis中获取，redis不存在重新获取，并更新到redis中
    token = rd.get('%s_token' % agentid)
    if token is None:
        token = get_token(wechat_corpsecret, agentid)
    # 记录一分钟发送消息条目，不超过30条
    timer = rd.get("timer")
    if timer is None:
        # 每次更新timer，就清空counter计数
        timer = int(time.time())
        # redis version 3.2
        try:
            rd.setex("timer", 60, timer)
        except Exception:
            rd.setex("timer", timer, 60)
        # redis version 3.0
        # rd.setex("timer", timer, 60)
        rd.set("counter", 0)
    timer = int(timer)
    counter = rd.get("counter")
    if counter is None:
        counter = 0
        rd.set("counter", counter)
    counter = int(counter)
    if counter > 30:  # 超过30条，暂停发送消息，等待下一分钟
        stime = 60 + timer - int(time.time())
        logger.warn("Sending messages exceeding 30 per person per minute limit, now start sleep %d seconds" % stime)
        time.sleep(stime)
    headers = {
        'Connection': 'keep-alive',
        'Content-Type': 'application/json; charset=utf-8'
    }
    hello = msg.splitlines()
    shit = []
    for i in range(len(hello)):
        if str(hello[i]).startswith("ITEM名称"):
            shit.append(i)
    msg_main_body = get_string_between_list_two_keys("ITEM值", "ITEM名称", hello, sep_str="\n")
    msg = "\n".join(hello[:13]) + "\n" + msg_main_body[:1200] + "\n" + "内容过长，只截取部分\n\n" + "\n".join(hello[shit[1]:])

    data = {
        "toparty": toparty,
        "totag": " ",
        "msgtype": "text",
        "agentid": agentid,
        "text": {
            "content": msg
        },
        "safe": "0"
    }
    params = {"access_token": token}
    try:
        response = requests.post(send_url, data=json.dumps(data, ensure_ascii=False), params=params,
                                 headers=headers, timeout=15)
        if response.status_code == 200:
            log_msg['code'] = 200
            log_msg['msg'] = msg
            try:
                logger.info(json.dumps(log_msg))
            except UnicodeDecodeError:
                logger.info("response code 200 msg:" + msg)
            counter += 1
            rd.set("counter", counter)
        else:
            raise Exception(response.json())
    except Exception as e:
        logger.error(str(e))


if __name__ == '__main__':
    logger = Logger()
    try:
        warn_dict = dict()
        msg = sys.argv[3]
        msg_list = msg.splitlines()
        hostname = msg_list[0].split()[1].strip("[]")
        warn_dict['hostname'] = hostname
        public_ip = 'a.b.c.d'
        level = "Other"
        status = "OK"
        trigger = None
        eventid = ''
        P = "P4"
        if "告警" in msg_list[0]:
            logger.error("zabbix alert level %s" % msg_list[2])
            eventid = msg_list[5].split(":")[1].strip()
            trigger = msg_list[1].split('||')[1].strip()
            level = msg_list[2].split(':')[1].strip()
            if alert_info.get(level, None):
                P = alert_info.get(level, None)["P"]
            msg = "\n".join(insert_into_list(msg_list, "事件编号", "公网IP: [%s]" % public_ip))
            msg += "\n\n告警时间：%s" % time_format['NOW_TIME']
            status = "PROBLEM"
            rd.set(eventid, time.time())
            p1_record(eventid, hostname, trigger, P, flag=False)
        elif "恢复" in msg_list[0]:
            eventid = msg_list[5].split(":")[1].strip()
            trigger = msg_list[1].split('||')[1].strip()
            level = msg_list[2].split(':')[1].strip()
            if alert_info.get(level, None):
                P = alert_info.get(level, None)["P"]
            recover_time = "恢复时间：%s" % time_format['NOW_TIME']
            msg_2 = insert_into_list(msg_list, "事件编号", "公网IP: [%s]" % public_ip)
            msg_2 = insert_into_list(msg_list, "事件编号", "%s" % recover_time)
            msg = "\n".join(msg_2)
            prev_time = rd.get(eventid)
            rd.delete(eventid)
            p1_record(eventid, hostname, trigger, P, flag=True)
            if prev_time is None:
                prev_time = time.time()
            time_used = float(time.time()) - float(prev_time)
            msg += "\n告警->恢复用时：%s" % convert_seconds_to_time(time_used)
            status = "OK"

        warn_dict['level'] = P
        warn_dict['status'] = status
        warn_dict['time'] = "%s" % time_format['NOW_TIME']
        warn_dict['trigger_name'] = trigger
        warn_dict['event_id'] = eventid
        logger.info(json.dumps(warn_dict))
        warn_1 = "yes"
        send_wechat(msg, level)
        if warn_1 == "yes":
            try:
                requests.get('%s?eventid=%s&action=alert&trigger=%s&level=%s' % (alert_url,
                                                                                 eventid, trigger, level), timeout=5)
            except Exception as e:
                # logger.error(str(e))
                pass
    except IOError as (errno, strerror):
        logger.error("I/O error({0}): {1}".format(errno, strerror))
    except ValueError:
        logger.error("Could not convert data to an integer.")
    except Exception as e:
        logger.error(str(e))
