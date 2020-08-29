#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/monitor_redis/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_redis.py"

[ $# -eq 0 ] && python ${script} -q || python ${script} -i ${1} -p ${2} ${3}