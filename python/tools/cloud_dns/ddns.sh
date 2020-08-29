#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/cloud_dns/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/ddns.py"
if [[ $# < 1 ]]; then
    exec python ${script} --help
    exit
fi
[ $# -eq 0 ] && python ${script} --help || python ${script} ${1} ${2} ${3} ${4}

