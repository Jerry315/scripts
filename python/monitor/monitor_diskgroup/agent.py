# -*- coding: utf-8 -*-
import json
import time
import socket
from common import AgentLogger, get_data, config


def agent(url):
    data = get_data(url)['monitor_info']
    for item in data:
        tmp = dict()
        tmp['timestamp'] = int(time.time())
        tmp['hostname'] = socket.gethostname()
        tmp['log_type'] = 'disc_rate'
        tmp['max_upload_rate'] = item['max_upload_rate']
        tmp['upload_rate'] = item['upload_rate']
        tmp['group_id'] = item['group_id']
        discs = []
        for disk in item['disk_infos']:
            # discs[disk["discID"]] = disk["upload_rate"]
            discs.append({"did": disk["discID"],"upload_rate": disk["upload_rate"]})
        tmp['discs'] = discs
        logger.log(json.dumps(tmp))


if __name__ == '__main__':
    logger = AgentLogger()
    try:
        url = config['codisktracker']['mointor_url']
        agent(url)
    except Exception as e:
        logger.log(str(e),mode=False)
