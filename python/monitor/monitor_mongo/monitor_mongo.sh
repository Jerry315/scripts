#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/monitor_mongo/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_mongodb.py"
if [[ $# < 1 ]]; then
    exec python ${script} query
    exit
fi
[ $# -eq 0 ] && python ${script} query || python ${script} ${1} ${2} ${3} ${4}

