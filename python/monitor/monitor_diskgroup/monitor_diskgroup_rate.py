# -*- coding: utf-8 -*-
import optparse
import sys
from common import logger, get_group_id, get_data, config


def get_group_upload_rate(gid, url):
    data = get_data(url)
    for item in data['monitor_info']:
        if item['group_id'] == gid:
            return item['upload_rate']


def main(method, url, gid=1):
    if method == 'group':
        result = get_group_id(url, get_data)
    elif method == 'rate':
        result = get_group_upload_rate(gid, url)
    else:
        result = 0
    print result


if __name__ == '__main__':
    try:
        parser = optparse.OptionParser()
        parser.add_option('-m', '--method', dest='method')
        parser.add_option('-g', '--gid', dest='group_id')
        argv = sys.argv[1:]
        (options, args) = parser.parse_args(argv)
        url = config['codisktracker']['mointor_url']
        method = options.method
        try:
            gid = int(options.group_id)
        except Exception:
            gid = 1
        main(method, url, gid)
    except Exception as e:
        logger.log(str(e),mode=False)
        print 0
