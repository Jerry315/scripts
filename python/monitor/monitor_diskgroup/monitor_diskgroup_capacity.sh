#!/usr/bin/env bash
export PYTHONUNBUFFERED=true
export PYTHONIOENCODING="utf-8"
source /var/virtualenvs/monitor_diskgroup/bin/activate

SCRIPT_PATH="$(cd "$(dirname "$0")"; pwd)"
script="${SCRIPT_PATH}/monitor_diskgroup_capacity.py"
if [[ $# < 1 ]]; then
    exec python ${script} --help
    exit
fi
ps_name=$*
[ $# -eq 0 ] && python ${script} --help || python ${script} ${ps_name}

